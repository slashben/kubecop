# Rules

## Rule list

| ID | Rule | Description | Tags | Priority | Application profile |
|----|------|-------------|------|----------|---------------------|
| R0001 | Exec Whitelisted | Detecting exec calls that are not whitelisted by application profile | [exec whitelisted] | 7 | true |
| R0002 | Unexpected file access | Detecting file access that are not whitelisted by application profile. File access is defined by the combination of path and flags | [open whitelisted] | 5 | true |
| R0003 | Unexpected system call | Detecting unexpected system calls that are not whitelisted by application profile. Every unexpected system call will be alerted only once. | [syscall whitelisted] | 7 | true |
| R0004 | Unexpected capability used | Detecting unexpected capabilities that are not whitelisted by application profile. Every unexpected capability is identified in context of a syscall and will be alerted only once per container. | [capabilities whitelisted] | 8 | true |
| R0005 | Unexpected domain request | Detecting unexpected domain requests that are not whitelisted by application profile. | [dns whitelisted] | 6 | true |
| R1000 | Exec from malicious source | Detecting exec calls that are from malicious source like: /dev/shm, /run, /var/run, /proc/self | [exec signature] | 9 | false |