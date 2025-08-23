package deck

import (
	"path/filepath"
	"github.com/joeblew999/infra/pkg/config"
)

// File and directory constants to prevent obfuscation
const (
	// Directory names (relative to data directory)
	DeckSubDir      = "deck"
	DeckCacheSubDir = "cache"
	DeckFontsSubDir = "fonts"
	
	// Output format constants
	FormatSVG = "svg"
	FormatPNG = "png"  
	FormatPDF = "pdf"
	FormatXML = "xml"
	
	// Default dimensions (Letter size in points)
	DefaultWidth  = 792.0
	DefaultHeight = 612.0
	
	// Layer rendering order
	DefaultLayers = "image:rect:ellipse:curve:arc:line:poly:text:list"
	
	// Font constants
	DefaultFontFamily = "Arial"
	DefaultFontWeight = 400
	DefaultFontSize   = 12.0
	
	// Color constants
	DefaultForeground = "black"
	DefaultBackground = "white"
	DefaultLineColor  = "rgb(127,127,127)"
	
	// Alignment constants
	AlignLeft   = "left"
	AlignCenter = "center" 
	AlignRight  = "right"
	AlignMiddle = "middle"
	AlignStart  = "start"
	AlignEnd    = "end"
	
	// List type constants
	ListTypeBullet = "bullet"
	ListTypeNumber = "number"
	
	// Common font paths (system fonts)
	SystemFontPathMacOS   = "/System/Library/Fonts/"
	SystemFontPathLinux   = "/usr/share/fonts/"
	SystemFontPathWindows = "C:\\Windows\\Fonts\\"
	
	// Font file names
	FontArialTTF      = "Arial.ttf"
	FontHelveticaTTF  = "Helvetica.ttf" 
	FontTimesNewTTF   = "Times New Roman.ttf"
	FontCourierTTF    = "Courier New.ttf"
	
	// Grid and spacing constants
	DefaultGridPercent   = 10.0
	DefaultLineSpacing   = 2.0
	DefaultStrokeWidth   = 2.0
	DefaultOpacity       = 1.0
	
	// Deck source constants (from original pngdeck/svgdeck/pdfdeck)
	MM2PT       = 2.83464 // mm to pt conversion
	LineSpacing = 1.4
	ListSpacing = 2.0
	FontFactor  = 1.0
	ListWrap    = 95.0
	
	// XML element names (for parsing)
	XMLElementDeck     = "deck"
	XMLElementSlide    = "slide" 
	XMLElementCanvas   = "canvas"
	XMLElementRect     = "rect"
	XMLElementEllipse  = "ellipse"
	XMLElementLine     = "line"
	XMLElementText     = "text"
	XMLElementList     = "list"
	XMLElementListItem = "li"
	XMLElementImage    = "image"
	
	// XML attribute names
	XMLAttrWidth    = "width"
	XMLAttrHeight   = "height"
	XMLAttrXP       = "xp"
	XMLAttrYP       = "yp"
	XMLAttrWP       = "wp"
	XMLAttrHP       = "hp"
	XMLAttrColor    = "color"
	XMLAttrOpacity  = "opacity"
	XMLAttrFont     = "font"
	XMLAttrAlign    = "align"
	XMLAttrType     = "type"
	XMLAttrScale    = "scale"
	XMLAttrSP       = "sp"
	XMLAttrLP       = "lp"
	XMLAttrBG       = "bg"
	XMLAttrFG       = "fg"
	XMLAttrXP1      = "xp1"
	XMLAttrYP1      = "yp1"
	XMLAttrXP2      = "xp2"
	XMLAttrYP2      = "yp2"
	XMLAttrName     = "name"
	
	// SVG constants
	SVGNamespace = "http://www.w3.org/2000/svg"
	SVGVersion   = "1.1"
	
	// File extensions
	ExtSVG = ".svg"
	ExtPNG = ".png"
	ExtPDF = ".pdf"
	ExtXML = ".xml"
	ExtDSH = ".dsh"
	
	// MIME types
	MimeTypeSVG = "image/svg+xml"
	MimeTypePNG = "image/png"
	MimeTypePDF = "application/pdf"
	MimeTypeXML = "application/xml"
	
	// Error messages
	ErrUnsupportedFormat    = "unsupported format"
	ErrFontManagementOff    = "font management is disabled"
	ErrSlideIndexOutOfRange = "slide index out of range"
	ErrNoSlidesFound        = "no slides found in deck"
	ErrInvalidDimensions    = "invalid canvas dimensions"
	ErrFileNotFound         = "file not found"
	ErrPermissionDenied     = "permission denied"
)

// Path functions using pkg/config

// GetDeckDataPath returns the absolute path to the deck data directory
func GetDeckDataPath() string {
	return filepath.Join(config.GetDataPath(), DeckSubDir)
}

// GetDeckCachePath returns the absolute path to the deck cache directory
func GetDeckCachePath() string {
	return filepath.Join(GetDeckDataPath(), DeckCacheSubDir)
}

// GetDeckFontPath returns the absolute path to the deck font cache directory
func GetDeckFontPath() string {
	return filepath.Join(GetDeckDataPath(), DeckFontsSubDir)
}