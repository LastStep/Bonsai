package guard

// VG round 2's M1, part 1 (15 Oct), its reproducers chmod2.py and the neighbours of its gap as tests: a change of
// permissions, owner, ACL or attributes onto Bonsai's own files is refused from every agent, and inside the project from
// roles that change nothing (attrwords.go).

import (
	"fmt"
	"path/filepath"
	"testing"
)

// M1, part 1 (chmod2.py and its neighbours): a change of permissions, owner, ACL or attributes onto Bonsai's own files,
// or onto a folder holding them, or by a glob that may name them, is refused as bonsai-files from every agent; the
// person's own session keeps today's guard.
func TestM1PermissionChangesOnBonsaiFiles(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:lead")
	calls := ".bonsai/local/agents/" + sessA + "/x.calls"
	bashLines := []string{
		// VG's.
		"chmod a-w .bonsai/local/agents/*/*.calls", "chmod 000 .bonsai/local/agents/" + sessA, "chmod -R a-w .bonsai/local/agents",
		"chown $USER .bonsai/local/agents/x.calls", "setfacl -m u::r .bonsai/local/agents/x.calls", "chmod a-w .bonsai/lock.json",
		"attrib +r .bonsai/local/agents/x.calls", "icacls .bonsai/local/agents /deny Everyone:W",
		// Every flag form (Git Bash takes a Windows program's path with forward slashes, or quoted backslashes).
		"chmod -w " + calls, "chmod -Rv u-w " + calls, "chmod --reference=free.txt " + calls, "chmod --ref free.txt " + calls,
		"chmod -v 444 -- " + calls, "chmod 0444 ./" + calls, "chgrp -R users .bonsai/local", "chown --from=a:b c:d " + calls,
		"chown --reference=free.txt " + calls, "chattr +i " + calls, "chattr -R -V =i .bonsai/local/agents", "chattr -v 2 -i " + calls,
		"setfacl --set=u::r,g::r,o::r " + calls, "setfacl -Rbm u::r .bonsai/local", "setfacl -M acl.txt " + calls,
		"setfattr -n user.x -v 1 " + calls, "chcon -t tmp_t " + calls, "chcon system_u:object_r:tmp_t:s0 " + calls,
		"chflags uchg " + calls, "sudo chmod 000 " + calls, "bash -c 'chmod 000 " + calls + "'", "git status && chmod 000 " + calls,
		`icacls .bonsai/local/agents //deny "Everyone:(W)"`, "icacls .bonsai/local/agents /grant:r Everyone:R /T",
		"icacls .bonsai/local /setowner nobody", "icacls .bonsai/local/agents /inheritance:r", "cacls .bonsai/local/agents/x.calls /E /D Everyone",
		"takeown /f .bonsai/local/agents/x.calls", "takeown //F .bonsai/local/agents //R", "attrib +r +h '.bonsai/local/*' //s //d",
		"icacls free.txt /save .bonsai/local/acl.txt",
		// A folder that holds them, and a glob that may name them.
		"chmod 000 .bonsai", "chmod a-x .bonsai/local", "chmod -R a-w .", "chmod a-w ..", "chmod 755 " + filepath.ToSlash(dir),
		"chmod a-w .bon*/local/agents/x.calls", "chmod 000 .bonsai/*/agents", "chmod 000 .bonsai/l?cal", "chown -R x: .[b]onsai",
		"cd .bonsai && chmod 000 local", "attrib +r .bonsai",
	}
	psLines := []string{
		"Set-ItemProperty -Path " + calls + " -Name IsReadOnly -Value $true", "sp " + calls + " IsReadOnly $true",
		`Set-ItemProperty -LiteralPath .bonsai\lock.json -Name Attributes -Value ReadOnly`,
		"(Get-Item " + calls + ").IsReadOnly = $true", "(gi " + calls + ").Attributes = 'ReadOnly'",
		"$(Get-Item -LiteralPath " + calls + ").IsReadOnly = $true", "(Get-Item " + calls + ").Attributes += 'ReadOnly'",
		"(Get-Item " + calls + ").IsReadOnly=$true", "(Get-Item " + calls + ").set_IsReadOnly($true)",
		"(Get-ChildItem .bonsai/local/agents).SetAccessControl($acl)", "Set-Acl -Path .bonsai/local/agents -AclObject $acl",
		"Set-Acl " + calls + " $acl", "Clear-ItemProperty -Path " + calls + " -Name Attributes", "clp " + calls + " Attributes",
		`attrib +r .bonsai\local\agents\x.calls`, `icacls .bonsai\local\agents /deny 'Everyone:(W)'`, "Set-ItemProperty .bonsai IsReadOnly $true",
	}
	for i, line := range bashLines {
		for _, typ := range []string{"workflow:lead", "workflow:helper", "Explore"} {
			wantAnswer(t, dir, shellAs(dir, fmt.Sprintf("m%s%d", typ[len(typ)-2:], i), typ, line, false), 2, RuleBonsaiFiles)
		}
	}
	for i, line := range psLines {
		for _, typ := range []string{"workflow:lead", "Explore"} {
			wantAnswer(t, dir, shellAs(dir, fmt.Sprintf("p%s%d", typ[len(typ)-2:], i), typ, line, true), 2, RuleBonsaiFiles)
		}
	}
	// The person's own session keeps today's guard there.
	wantAnswer(t, dir, shellAs(dir, "", "", "chmod a-w "+calls, false), 0, "")
	wantAnswer(t, dir, shellAs(dir, "", "", "(Get-Item "+calls+").IsReadOnly = $true", true), 0, "")
}

// M1, part 1: inside the project, a role that changes nothing is refused every such change (role-writes-nothing), and
// may still read attributes and ACLs; a role that writes changes its own files' modes as before.
func TestM1PermissionChangesByRole(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:lead")
	tmp := filepath.ToSlash(t.TempDir())
	changes := []string{"chmod +x free.txt", "chmod -R 755 src", "chown x free.txt", "chgrp users free.txt", "setfacl -m u::r free.txt",
		"chattr +i free.txt", "attrib +r free.txt", "icacls free.txt /grant Everyone:F", "takeown /f free.txt", "cacls free.txt /E /G x:R"}
	psChanges := []string{"Set-ItemProperty free.txt IsReadOnly $true", "(gi free.txt).IsReadOnly = $true", "Set-Acl free.txt $acl"}
	for i, line := range changes {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("h", i), "workflow:helper", line, false), 2, RuleRoleWritesNothing)
	}
	for i, line := range psChanges {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("hp", i), "workflow:helper", line, true), 2, RuleRoleWritesNothing)
	}
	reads := []string{"attrib free.txt", "icacls free.txt", "cacls free.txt", "chmod --help", "ls -l .bonsai/local/agents",
		"chmod +x " + tmp + "/build.sh", "stat -c %a free.txt"}
	for i, line := range reads {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("r", i), "workflow:helper", line, false), 0, "")
	}
	for i, line := range []string{"Get-Acl free.txt", "Write-Output (Get-Item free.txt).IsReadOnly", "(Get-Item free.txt).Length"} {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("rp", i), "workflow:helper", line, true), 0, "")
	}
	// A lead (a role that writes) changes its own files' modes; a mode in a variable is not read (the known limit).
	for _, line := range append(changes[:5:5], "chmod u+x \"$f\"", "chmod 644 *", "chmod -R u+w src/*") {
		wantAnswer(t, dir, shellAs(dir, "wl", "workflow:lead", line, false), 0, "")
	}
	for _, line := range psChanges {
		wantAnswer(t, dir, shellAs(dir, "wl", "workflow:lead", line, true), 0, "")
	}
}
