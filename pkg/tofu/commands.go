package tofu

// Init runs tofu init
func (r *Runner) Init() error {
	return r.Run("init")
}

// Plan runs tofu plan
func (r *Runner) Plan() error {
	return r.Run("plan")
}

// Apply runs tofu apply with auto-approve
func (r *Runner) Apply() error {
	return r.Run("apply", "-auto-approve")
}

// ApplyInteractive runs tofu apply without auto-approve
func (r *Runner) ApplyInteractive() error {
	return r.Run("apply")
}

// Destroy runs tofu destroy with auto-approve
func (r *Runner) Destroy() error {
	return r.Run("destroy", "-auto-approve")
}

// DestroyInteractive runs tofu destroy without auto-approve
func (r *Runner) DestroyInteractive() error {
	return r.Run("destroy")
}

// Validate runs tofu validate
func (r *Runner) Validate() error {
	return r.Run("validate")
}

// Version gets tofu version
func (r *Runner) Version() ([]byte, error) {
	return r.RunWithOutput("version")
}
