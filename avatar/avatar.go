package avatar

import (
	"bytes"
	"errors"
	"image/color"
	"image/png"
	"strings"
	"unicode"

	"github.com/dchest/lru"
	"stathat.com/c/consistent"
)

var (
	avatarBgColors = map[string]*color.RGBA{
		"45BDF3": &color.RGBA{69, 189, 243, 255},
		"E08F70": &color.RGBA{224, 143, 112, 255},
		"4DB6AC": &color.RGBA{77, 182, 172, 255},
		"9575CD": &color.RGBA{149, 117, 205, 255},
		"B0855E": &color.RGBA{176, 133, 94, 255},
		"F06292": &color.RGBA{240, 98, 146, 255},
		"A3D36C": &color.RGBA{163, 211, 108, 255},
		"7986CB": &color.RGBA{121, 134, 203, 255},
		"F1B91D": &color.RGBA{241, 185, 29, 255},
	}

	defaultColorKey = "45BDF3"

	// ErrUnsupportChar is returned when the character is not supported
	ErrUnsupportChar = errors.New("unsupported character")

	c = consistent.New()
)

// InitialsAvatar represents an initials avatar.
type InitialsAvatar struct {
	drawer *drawer
	cache  *lru.Cache
}

// New creates an instance of InitialsAvatar
func New(fontFile string) *InitialsAvatar {
	avatar := NewWithConfig(Config{
		MaxItems: 1024, // default to 1024 items.
		FontFile: fontFile,
	})
	return avatar
}

// Config is the configuration object for avatar images.
type Config struct {
	// Maximum number of items the cache can contain (unlimited by default).
	MaxItems int

	// Maximum byte capacity of cache (unlimited by default).
	MaxBytes int64

	// TrueType Font file path
	FontFile string
}

// NewWithConfig provides config for LRU Cache.
func NewWithConfig(cfg Config) *InitialsAvatar {
	avatar := new(InitialsAvatar)
	avatar.drawer = newDrawer(cfg.FontFile)
	avatar.cache = lru.New(lru.Config{
		MaxItems: cfg.MaxItems,
		MaxBytes: cfg.MaxBytes,
	})

	return avatar
}

// DrawToBytes draws an image base on the name and size.
// Only initials of name will be draw.
// The size is the side length of the square image. Image is encoded to bytes.
func (a *InitialsAvatar) DrawToBytes(name string, size int) ([]byte, error) {
	if size <= 0 {
		size = 48 // default size
	}
	name = strings.TrimSpace(name)
	firstRune := []rune(name)[0]
	if !isHan(firstRune) && !unicode.IsLetter(firstRune) {
		return nil, ErrUnsupportChar
	}
	initials := getInitials(name)
	bgcolor := getColorByName(name)

	// get from cache
	v, ok := a.cache.GetBytes(lru.Key(initials))
	if ok {
		return v, nil
	}

	m := a.drawer.Draw(initials, size, bgcolor)

	var buf bytes.Buffer
	err := png.Encode(&buf, m)
	if err != nil {
		return nil, err
	}
	// set cache
	a.cache.SetBytes(lru.Key(initials), buf.Bytes())

	return buf.Bytes(), nil
}

// is Chinese?
func isHan(r rune) bool {
	if unicode.Is(unicode.Scripts["Han"], r) {
		return true
	}
	return false
}

// random color
func getColorByName(name string) *color.RGBA {
	key, err := c.Get(name)
	if err != nil {
		key = defaultColorKey
	}
	return avatarBgColors[key]
}

//TODO: enhance
func getInitials(name string) string {
	if len(name) <= 0 {
		return ""
	}
	return strings.ToUpper(string([]rune(name)[0]))
}

func init() {
	for key := range avatarBgColors {
		c.Add(key)
	}
}
