
<a name="v0.1.0"></a>
## v0.1.0 - 2026-10-28
### Features
- implement profile command and subcommands (WORKBENCH-4405)
- support profiles in the config file (WORKBENCH-4455)
- implement a whoami command (WORKBENCH-4484)
- implement a verbose option (WORKBENCH-4482)
- add version command (WORKBENCH-4475)
- implement a logout command (WORKBENCH-4477)
- remove password and client credentials auth modes (WORKBENCH-4487)
- add alases for job-submission create and cancel (WORKBENCH-4492)
- implement device auth via redirection to authorization endpiont (WORKBENCH-4466)
- support mounts and capabilities (WORKBENCH-4445)
- initial commit

### Bug Fixes
- improve the output of the version command (WORKBENCH-4475)
- restore client credentials auth mode (WORKBENCH-4487)
- normalize endpoint url in token cache (WORKBENCH-4486)
- show device code in device auth output (WORKBENCH-4479)
- require organization when performing device auth (WORKBENCH-4485)
- include applications, services, and nodepools in error messages (WORKBENCH-4489)
- pass organization in scope
- job creation payload now includes properites (WORKBENCH-4472)

### Documentation
- update readme to remove client-id references (WORKBENCH-4478)
- update usage for auth mode changes (WORKBENCH-4487)

### Build Changes
- **deps:** bump github.com/kubernetes-csi/external-snapshotter/client/v8
- **deps:** bump the gomod group with 4 updates

### Continuous Integration
- update text in gen-usage.sh


[Unreleased]: https://github.com/sassoftware/wbctl/compare/v0.1.0...HEAD
