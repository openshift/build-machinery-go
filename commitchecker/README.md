# commitchecker

Cmdline tool that checks all commits in git branch or upstream PR have
[format required by OpenShift for cherry-picks](https://github.com/openshift/kubernetes/blob/master/README.openshift.md#cherry-picking-an-upstream-change).

## Usage

```
Usage of ./commitchecker:
  -end string
	The end of the revision range for analysis (default "HEAD")
  -start string
	The start of the revision range for analysis (default "master")
```
