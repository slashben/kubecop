package rule

import (
	"fmt"

	"github.com/armosec/kubecop/pkg/approfilecache"
	"github.com/kubescape/kapprofiler/pkg/tracing"
)

const (
	R0001ExecWhitelistedRuleName = "R-0001 Exec Whitelisted"
)

var R0001ExecWhitelistedRuleDescriptor = RuleDesciptor{
	Name: R0001ExecWhitelistedRuleName,
	Tags: []string{"exec", "whitelisted"},
	Requirements: RuleRequirements{
		EventTypes:             []tracing.EventType{tracing.ExecveEventType},
		NeedApplicationProfile: true,
	},
	RuleCreationFunc: func() Rule {
		return CreateRuleR0001ExecWhitelisted()
	},
}

type R0001ExecWhitelisted struct {
}

type R0001ExecWhitelistedFailure struct {
	RuleName     string
	Err          string
	FailureEvent *tracing.ExecveEvent
}

func (rule *R0001ExecWhitelisted) Name() string {
	return R0001ExecWhitelistedRuleName
}

func CreateRuleR0001ExecWhitelisted() *R0001ExecWhitelisted {
	return &R0001ExecWhitelisted{}
}

func (rule *R0001ExecWhitelisted) DeleteRule() {
}

func (rule *R0001ExecWhitelisted) ProcessEvent(eventType tracing.EventType, event interface{}, appProfileAccess approfilecache.SingleApplicationProfileAccess) RuleFailure {
	if eventType != tracing.ExecveEventType {
		return nil
	}

	execEvent, ok := event.(*tracing.ExecveEvent)
	if !ok {
		return nil
	}

	appProfileExecList, err := appProfileAccess.GetExecList()
	if err != nil || appProfileExecList == nil {
		return &R0001ExecWhitelistedFailure{
			RuleName:     rule.Name(),
			Err:          "Application profile is missing",
			FailureEvent: execEvent,
		}
	}

	for _, execCall := range *appProfileExecList {
		if execCall.Path == execEvent.PathName {
			return nil
		}
	}

	return &R0001ExecWhitelistedFailure{
		RuleName:     rule.Name(),
		Err:          fmt.Sprintf("exec call \"%s\" is not whitelisted by application profile", execEvent.PathName),
		FailureEvent: execEvent,
	}
}

func (rule *R0001ExecWhitelisted) Requirements() RuleRequirements {
	return RuleRequirements{
		EventTypes:             []tracing.EventType{tracing.ExecveEventType},
		NeedApplicationProfile: true,
	}
}

func (rule *R0001ExecWhitelistedFailure) Name() string {
	return rule.RuleName
}

func (rule *R0001ExecWhitelistedFailure) Error() string {
	return rule.Err
}

func (rule *R0001ExecWhitelistedFailure) Event() tracing.GeneralEvent {
	return rule.FailureEvent.GeneralEvent
}
