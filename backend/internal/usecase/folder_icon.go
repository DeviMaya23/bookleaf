package usecase

// folderIconAllowlist is the fixed set of icon keys a folder's Icon field may
// hold. Keys correspond to lucide-react icon components; the frontend
// maintains its own hand-written mirror of this list (see design.md).
var folderIconAllowlist = map[string]struct{}{
	"bookmark":           {},
	"paperclip":          {},
	"folder":             {},
	"folder-bookmark":    {},
	"folder-closed":      {},
	"folder-open":        {},
	"folders":            {},
	"file-stack":         {},
	"file-question-mark": {},
	"file-image":         {},
	"book-image":         {},
	"image":              {},
	"images":             {},
	"gpu":                {},
	"mirror-rectangular": {},
	"sun":                {},
	"moon":               {},
	"cloud":              {},
	"cloud-fog":          {},
	"cloud-drizzle":      {},
	"cloud-sun":          {},
	"cloudy":             {},
	"apple":              {},
	"coffee":             {},
	"cookie":             {},
	"chef-hat":           {},
	"sandwich":           {},
	"bottle-wine":        {},
	"clover":             {},
	"club":               {},
	"crown":              {},
	"gem":                {},
	"gift":               {},
	"headphones":         {},
	"rocket":             {},
	"star":               {},
	"ghost":              {},
	"house":              {},
	"heart":              {},
	"flower":             {},
	"leaf":               {},
	"sprout":             {},
	"trees":              {},
	"map-pin":            {},
	"utensils":           {},
	"ship-wheel":         {},
	"bell":               {},
	"alarm-clock":        {},
	"album":              {},
	"flask-conical":      {},
	"snowflake":          {},
	"cylinder":           {},
	"mail":               {},
	"palette":            {},
	"trash-2":            {},
}

// IsValidFolderIcon reports whether key is a member of the folder icon allowlist.
func IsValidFolderIcon(key string) bool {
	_, ok := folderIconAllowlist[key]
	return ok
}
