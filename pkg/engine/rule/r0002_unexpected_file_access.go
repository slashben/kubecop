package rule

import (
	"fmt"

	"github.com/armosec/kubecop/pkg/approfilecache"
	"github.com/kubescape/kapprofiler/pkg/tracing"
)

const (
	R0002ID                           = "R0002"
	R0002UnexpectedFileAccessRuleName = "Unexpected file access"
)

var R0002UnexpectedFileAccessRuleDescriptor = RuleDesciptor{
	ID:          R0002ID,
	Name:        R0002UnexpectedFileAccessRuleName,
	Description: "Detecting file access that are not whitelisted by application profile. File access is defined by the combination of path and flags",
	Tags:        []string{"open", "whitelisted"},
	Priority:    5,
	Requirements: RuleRequirements{
		EventTypes:             []tracing.EventType{tracing.OpenEventType},
		NeedApplicationProfile: true,
	},
	RuleCreationFunc: func() Rule {
		return CreateRuleR0002UnexpectedFileAccess()
	},
}

type R0002UnexpectedFileAccess struct {
}

type R0002UnexpectedFileAccessFailure struct {
	RuleName         string
	RulePriority     int
	Err              string
	FixSuggestionMsg string
	FailureEvent     *tracing.OpenEvent
}

func (rule *R0002UnexpectedFileAccess) Name() string {
	return R0002UnexpectedFileAccessRuleName
}

func CreateRuleR0002UnexpectedFileAccess() *R0002UnexpectedFileAccess {
	return &R0002UnexpectedFileAccess{}
}

func (rule *R0002UnexpectedFileAccess) DeleteRule() {
}

func (rule *R0002UnexpectedFileAccess) generatePatchCommand(event *tracing.OpenEvent, appProfileAccess approfilecache.SingleApplicationProfileAccess) string {
	flagList := "["
	for _, arg := range event.Flags {
		flagList += "\"" + arg + "\","
	}
	// remove the last comma
	if len(flagList) > 1 {
		flagList = flagList[:len(flagList)-1]
	}
	baseTemplate := "kubectl patch applicationprofile %s --namespace %s --type merge -p '{\"spec\": {\"containers\": [{\"name\": \"%s\", \"opens\": [{\"path\": \"%s\", \"flags\": %s}]}]}}'"
	return fmt.Sprintf(baseTemplate, appProfileAccess.GetName(), appProfileAccess.GetNamespace(),
		event.ContainerName, event.PathName, flagList)
}

func (rule *R0002UnexpectedFileAccess) ProcessEvent(eventType tracing.EventType, event interface{}, appProfileAccess approfilecache.SingleApplicationProfileAccess, engineAccess EngineAccess) RuleFailure {
	if eventType != tracing.OpenEventType {
		return nil
	}

	openEvent, ok := event.(*tracing.OpenEvent)
	if !ok {
		return nil
	}

	if appProfileAccess == nil {
		return &R0002UnexpectedFileAccessFailure{
			RuleName:         rule.Name(),
			Err:              "Application profile is missing",
			FixSuggestionMsg: fmt.Sprintf("Please create an application profile for the Pod %s", openEvent.PodName),
			FailureEvent:     openEvent,
			RulePriority:     RulePrioritySystemIssue,
		}
	}

	appProfileOpenList, err := appProfileAccess.GetOpenList()
	if err != nil || appProfileOpenList == nil {
		return &R0002UnexpectedFileAccessFailure{
			RuleName:         rule.Name(),
			Err:              "Application profile is missing",
			FixSuggestionMsg: fmt.Sprintf("Please create an application profile for the Pod %s", openEvent.PodName),
			FailureEvent:     openEvent,
			RulePriority:     RulePrioritySystemIssue,
		}
	}

	for _, open := range *appProfileOpenList {
		if open.Path == openEvent.PathName {
			found := 0
			for _, eventOpenFlag := range openEvent.Flags {
				// Check that event open flag is in the open.Flags
				for _, profileOpenFlag := range open.Flags {
					if eventOpenFlag == profileOpenFlag {
						found += 1
					}
				}
			}
			if found == len(openEvent.Flags) {
				return nil
			}
			// TODO: optimize this list (so path will be only once in the list so we can break the loop)
		}
	}

	return &R0002UnexpectedFileAccessFailure{
		RuleName:         rule.Name(),
		Err:              fmt.Sprintf("Unexpected file access: %s with flags %v", openEvent.PathName, openEvent.Flags),
		FixSuggestionMsg: fmt.Sprintf("If this is a valid behavior, please add the open call \"%s\" to the whitelist in the application profile for the Pod \"%s\". You can use the following command: %s", openEvent.PathName, openEvent.PodName, rule.generatePatchCommand(openEvent, appProfileAccess)),
		FailureEvent:     openEvent,
		RulePriority:     R0002UnexpectedFileAccessRuleDescriptor.Priority,
	}
}

func (rule *R0002UnexpectedFileAccess) Requirements() RuleRequirements {
	return RuleRequirements{
		EventTypes:             []tracing.EventType{tracing.OpenEventType},
		NeedApplicationProfile: true,
	}
}

func (rule *R0002UnexpectedFileAccessFailure) Name() string {
	return rule.RuleName
}

func (rule *R0002UnexpectedFileAccessFailure) Error() string {
	return rule.Err
}

func (rule *R0002UnexpectedFileAccessFailure) Event() tracing.GeneralEvent {
	return rule.FailureEvent.GeneralEvent
}

func (rule *R0002UnexpectedFileAccessFailure) Priority() int {
	return rule.RulePriority
}

func (rule *R0002UnexpectedFileAccessFailure) FixSuggestion() string {
	return rule.FixSuggestionMsg
}
