package setup

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSetupLocalDarwinStartsColimaWhenServerIsMissing(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner(map[string]bool{
		"brew":   true,
		"colima": true,
		"incus":  true,
	})
	runner.queue("incus version", fakeRunResult{
		stdout: "Client version: 6.10",
		stderr: "Error: failed to connect to the Incus server",
		err:    errors.New("exit status 1"),
	})
	runner.queue("colima start --runtime incus --cpu 4 --memory 8 --nested-virtualization --vm-type vz", fakeRunResult{})
	runner.queue("incus version", fakeRunResult{
		stdout: "Client version: 6.10\nServer version: 6.10",
	})

	service := &Service{
		prompt:       &fakePrompter{},
		runner:       runner,
		hostDetector: fakeHostDetector{host: Host{GOOS: "darwin", Hostname: "test-client"}},
		out:          io.Discard,
		scriptWriter: func() (string, func(), error) { return "/tmp/unused", func() {}, nil },
	}

	if err := service.setupLocalDarwin(context.Background()); err != nil {
		t.Fatalf("setupLocalDarwin returned an error: %v", err)
	}

	if !runner.called("colima start --runtime incus --cpu 4 --memory 8 --nested-virtualization --vm-type vz") {
		t.Fatal("expected Colima start to be invoked")
	}
}

func TestSetupLocalDarwinUpdatesColimaWhenVersionsMismatch(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner(map[string]bool{
		"brew":   true,
		"colima": true,
		"incus":  true,
	})
	runner.queue("incus version", fakeRunResult{
		stdout: "Client version: 6.10\nServer version: 6.9",
	})
	runner.queue("colima update", fakeRunResult{})
	runner.queue("incus version", fakeRunResult{
		stdout: "Client version: 6.10\nServer version: 6.10",
	})

	service := &Service{
		prompt:       &fakePrompter{},
		runner:       runner,
		hostDetector: fakeHostDetector{host: Host{GOOS: "darwin", Hostname: "test-client"}},
		out:          io.Discard,
		scriptWriter: func() (string, func(), error) { return "/tmp/unused", func() {}, nil },
	}

	if err := service.setupLocalDarwin(context.Background()); err != nil {
		t.Fatalf("setupLocalDarwin returned an error: %v", err)
	}

	if !runner.called("colima update") {
		t.Fatal("expected Colima update to be invoked")
	}
}

func TestRunRemoteInstallsServerAndAddsRemote(t *testing.T) {
	t.Parallel()

	prompt := &fakePrompter{
		selectChoices: []int{1},
		askAnswers: []string{
			"root@198.51.100.10",
			"198.51.100.10",
		},
	}

	runner := newFakeRunner(map[string]bool{
		"incus": true,
		"scp":   true,
		"ssh":   true,
	})

	precheck := `if [ "$(id -u)" -eq 0 ] || sudo -n true >/dev/null 2>&1; then printf ok; else exit 1; fi`
	installSnippet := `chmod +x '/tmp/capsule-incus.abcd' && if [ "$(id -u)" -eq 0 ]; then '/tmp/capsule-incus.abcd' --mode='server'; else sudo '/tmp/capsule-incus.abcd' --mode='server'; fi; status=$?; rm -f '/tmp/capsule-incus.abcd'; exit $status`
	trustSnippet := `if [ "$(id -u)" -eq 0 ]; then incus config trust add 'test-client'; else sudo incus config trust add 'test-client'; fi`

	runner.queue(`incus remote list --format=csv`, fakeRunResult{})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(precheck), fakeRunResult{stdout: "ok"})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 mktemp /tmp/capsule-incus.XXXXXX", fakeRunResult{stdout: "/tmp/capsule-incus.abcd\n"})
	runner.queue("scp -o BatchMode=yes -o StrictHostKeyChecking=accept-new /tmp/capsule-script.sh root@198.51.100.10:/tmp/capsule-incus.abcd", fakeRunResult{})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(installSnippet), fakeRunResult{})
	runner.queue(`incus remote list --format=csv`, fakeRunResult{})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(trustSnippet), fakeRunResult{stdout: "token-value\n"})
	runner.queue(`incus remote add capsule https://198.51.100.10:8443 --accept-certificate --token token-value`, fakeRunResult{})
	runner.queue(`incus list capsule:`, fakeRunResult{stdout: ""})
	runner.queue(`incus remote switch capsule`, fakeRunResult{})

	service := &Service{
		prompt:       prompt,
		runner:       runner,
		hostDetector: fakeHostDetector{host: Host{GOOS: "darwin", Hostname: "test-client"}},
		out:          io.Discard,
		scriptWriter: func() (string, func(), error) { return "/tmp/capsule-script.sh", func() {}, nil },
	}

	if err := service.Run(context.Background()); err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}

	if !runner.called("incus remote add capsule https://198.51.100.10:8443 --accept-certificate --token token-value") {
		t.Fatal("expected remote add to be invoked")
	}
	if !runner.called("incus remote switch capsule") {
		t.Fatal("expected default remote switch to be invoked")
	}
	if len(prompt.askQuestions) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompt.askQuestions))
	}
	if !strings.Contains(prompt.askQuestions[0], "root@203.0.113.10") {
		t.Fatalf("expected SSH target prompt to include an example, got %q", prompt.askQuestions[0])
	}
}

func TestRunRemotePromptsForRemoteNameWhenCapsuleExists(t *testing.T) {
	t.Parallel()

	prompt := &fakePrompter{
		selectChoices: []int{1},
		askAnswers: []string{
			"root@198.51.100.10",
			"198.51.100.10",
			"lab",
		},
	}

	runner := newFakeRunner(map[string]bool{
		"incus": true,
		"scp":   true,
		"ssh":   true,
	})

	precheck := `if [ "$(id -u)" -eq 0 ] || sudo -n true >/dev/null 2>&1; then printf ok; else exit 1; fi`
	installSnippet := `chmod +x '/tmp/capsule-incus.abcd' && if [ "$(id -u)" -eq 0 ]; then '/tmp/capsule-incus.abcd' --mode='server'; else sudo '/tmp/capsule-incus.abcd' --mode='server'; fi; status=$?; rm -f '/tmp/capsule-incus.abcd'; exit $status`
	trustSnippet := `if [ "$(id -u)" -eq 0 ]; then incus config trust add 'test-client'; else sudo incus config trust add 'test-client'; fi`

	runner.queue(`incus remote list --format=csv`, fakeRunResult{stdout: "capsule,https://198.51.100.10:8443\n"})
	runner.queue(`incus remote list --format=csv`, fakeRunResult{stdout: "capsule,https://198.51.100.10:8443\n"})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(precheck), fakeRunResult{stdout: "ok"})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 mktemp /tmp/capsule-incus.XXXXXX", fakeRunResult{stdout: "/tmp/capsule-incus.abcd\n"})
	runner.queue("scp -o BatchMode=yes -o StrictHostKeyChecking=accept-new /tmp/capsule-script.sh root@198.51.100.10:/tmp/capsule-incus.abcd", fakeRunResult{})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(installSnippet), fakeRunResult{})
	runner.queue(`incus remote list --format=csv`, fakeRunResult{stdout: "capsule,https://198.51.100.10:8443\n"})
	runner.queue("ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@198.51.100.10 "+remoteShell(trustSnippet), fakeRunResult{stdout: "token-value\n"})
	runner.queue(`incus remote add lab https://198.51.100.10:8443 --accept-certificate --token token-value`, fakeRunResult{})
	runner.queue(`incus list lab:`, fakeRunResult{stdout: ""})
	runner.queue(`incus remote switch lab`, fakeRunResult{})

	service := &Service{
		prompt:       prompt,
		runner:       runner,
		hostDetector: fakeHostDetector{host: Host{GOOS: "darwin", Hostname: "test-client"}},
		out:          io.Discard,
		scriptWriter: func() (string, func(), error) { return "/tmp/capsule-script.sh", func() {}, nil },
	}

	if err := service.Run(context.Background()); err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}

	if !runner.called("incus remote add lab https://198.51.100.10:8443 --accept-certificate --token token-value") {
		t.Fatal("expected alternate remote add to be invoked")
	}
	if prompt.askDefaults[2] != "capsule-198-51-100-10" {
		t.Fatalf("expected alternate remote name suggestion, got %q", prompt.askDefaults[2])
	}
}

type fakePrompter struct {
	selectChoices []int
	askAnswers    []string
	askQuestions  []string
	askDefaults   []string
	confirmAnswer bool
}

func (f *fakePrompter) Select(_ string, _ []string) (int, error) {
	if len(f.selectChoices) == 0 {
		return 0, nil
	}

	answer := f.selectChoices[0]
	f.selectChoices = f.selectChoices[1:]
	return answer, nil
}

func (f *fakePrompter) Confirm(_ string, _ bool) (bool, error) {
	return f.confirmAnswer, nil
}

func (f *fakePrompter) Ask(question, defaultValue string) (string, error) {
	f.askQuestions = append(f.askQuestions, question)
	f.askDefaults = append(f.askDefaults, defaultValue)

	if len(f.askAnswers) == 0 {
		return defaultValue, nil
	}

	answer := f.askAnswers[0]
	f.askAnswers = f.askAnswers[1:]
	if answer == "" {
		return defaultValue, nil
	}

	return answer, nil
}

type fakeHostDetector struct {
	host Host
	err  error
}

func (f fakeHostDetector) Detect() (Host, error) {
	return f.host, f.err
}

type fakeRunResult struct {
	stdout string
	stderr string
	err    error
}

type fakeRunner struct {
	lookPath map[string]bool
	queued   map[string][]fakeRunResult
	calls    []string
}

func newFakeRunner(lookPath map[string]bool) *fakeRunner {
	return &fakeRunner{
		lookPath: lookPath,
		queued:   map[string][]fakeRunResult{},
	}
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if f.lookPath[file] {
		return "/usr/bin/" + file, nil
	}

	return "", errors.New("not found")
}

func (f *fakeRunner) Run(_ context.Context, spec CommandSpec) (Result, error) {
	key := strings.Join(append([]string{spec.Name}, spec.Args...), " ")
	f.calls = append(f.calls, key)

	queue := f.queued[key]
	if len(queue) == 0 {
		return Result{}, errors.New("unexpected command: " + key)
	}

	result := queue[0]
	f.queued[key] = queue[1:]

	return Result{
		Stdout: result.stdout,
		Stderr: result.stderr,
	}, result.err
}

func (f *fakeRunner) queue(command string, result fakeRunResult) {
	f.queued[command] = append(f.queued[command], result)
}

func (f *fakeRunner) called(command string) bool {
	for _, call := range f.calls {
		if call == command {
			return true
		}
	}

	return false
}
