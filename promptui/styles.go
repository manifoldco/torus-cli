package promptui

var (
	bold       = styler(fgBold)
	faint      = styler(fgFaint)
	underlined = styler(fgUnderline)
)

var (
	iconInitial = styler(fgBlue)("?")
	iconGood    = styler(fgGreen)("✔")
	iconWarn    = styler(fgYellow)("⚠")
	iconBad     = styler(fgRed)("✗")
)

var red = styler(fgBold, fgRed)
