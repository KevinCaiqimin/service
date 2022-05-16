package login

func init() {

}

type LoginMod struct {
}

func (l *LoginMod) Login(acc string, pwd string) bool {
	return true
}
