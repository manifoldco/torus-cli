package promptui

func styler(attrs ...attribute) func(string) string {
	return func(s string) string {
		return s
	}
}
