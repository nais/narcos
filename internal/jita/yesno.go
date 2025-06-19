package jita

type YesNoIcon bool

func (yn YesNoIcon) String() string {
	if yn {
		return "✅"
	} else {
		return "⛔"
	}
}
