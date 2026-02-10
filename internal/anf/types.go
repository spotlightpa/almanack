package anf

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"
)

func cloneOrInitMap[M map[K]V, K comparable, V any](m M) M {
	if m == nil {
		return make(M)
	}
	return maps.Clone(m)
}

func (a Article) Clone() Article {
	a.ComponentTextStyles = cloneOrInitMap(a.ComponentTextStyles)
	a.TextStyles = cloneOrInitMap(a.TextStyles)
	a.ComponentLayouts = cloneOrInitMap(a.ComponentLayouts)
	a.ComponentStyles = cloneOrInitMap(a.ComponentStyles)
	return a
}

// Article represents the root structure of an Apple News Format document
type Article struct {
	Version             string                        `json:"version"`
	Identifier          string                        `json:"identifier"`
	Title               string                        `json:"title"`
	Language            string                        `json:"language"`
	Layout              Layout                        `json:"layout"`
	Components          Components                    `json:"components"`
	Subtitle            string                        `json:"subtitle,omitempty"`
	AdvertisingSettings *AdvertisingSettings          `json:"advertisingSettings,omitempty"`
	Metadata            *Metadata                     `json:"metadata,omitempty"`
	DocumentStyle       *DocumentStyle                `json:"documentStyle,omitempty"`
	ComponentTextStyles map[string]ComponentTextStyle `json:"componentTextStyles"`
	TextStyles          map[string]TextStyle          `json:"textStyles"`
	ComponentLayouts    map[string]ComponentLayout    `json:"componentLayouts"`
	ComponentStyles     map[string]ComponentStyle     `json:"componentStyles"`
}

type Components []Component

func (comps *Components) UnmarshalJSON(data []byte) error {
	var rawComponents []json.RawMessage
	if err := json.Unmarshal(data, &rawComponents); err != nil {
		return err
	}

	*comps = make(Components, len(rawComponents))

	for i, rawComponent := range rawComponents {
		var roleExtractor struct {
			Role string `json:"role"`
		}

		if err := json.Unmarshal(rawComponent, &roleExtractor); err != nil {
			return err
		}

		var component Component
		switch roleExtractor.Role {
		case "author", "body", "byline", "caption", "heading", "heading1", "heading2", "heading3", "heading4", "heading5", "heading6", "illustrator", "intro", "photographer", "pullquote", "quote", "title":
			component = &TextComponent{}
		// case "audio", "music":
		// 	component = &AudioComponent{}
		// case "container":
		// 	component = &ContainerComponent{}
		// case "section":
		// 	component = &SectionComponent{}
		// case "chapter":
		// 	component = &ChapterComponent{}
		case "header":
			tc := &TextComponent{}
			tc.Role = "header"
			component = tc
		case "divider":
			component = &DividerComponent{}
		// case "embedwebvideo", "embedvideo":
		// 	component = &EmbedWebVideoComponent{}
		case "video":
			component = &VideoComponent{}
		// case "facebook_post":
		// 	component = &FacebookPostComponent{}
		// case "instagram":
		// 	component = &InstagramComponent{}
		// case "tweet":
		// 	component = &TweetComponent{}
		case "image", "logo", "figure":
			component = &ImageComponent{}
		case "photo", "portrait":
			component = &PhotoComponent{}
		case "gallery", "mosaic":
			component = &GalleryComponent{}
		case "map":
			component = &MapComponent{}
		// case "htmltable":
		// 	component = &HTMLTableComponent{}
		case "datatable":
			component = &DataTableComponent{}
		// case "arkit":
		// 	component = &ARKitComponent{}
		// case "banner_advertisement":
		// 	component = &BannerAdvertisementComponent{}
		// case "medium_rectangle_advertisement":
		// 	component = &MediumRectangleAdvertisementComponent{}
		default:
			return fmt.Errorf("unknown component role: %s", roleExtractor.Role)
		}

		if err := json.Unmarshal(rawComponent, component); err != nil {
			return err
		}

		(*comps)[i] = component
	}

	return nil
}

// Layout defines the article's column system
type Layout struct {
	Columns int `json:"columns"`
	Width   int `json:"width"`
	Gutter  int `json:"gutter,omitempty"` // default: 20
	Margin  int `json:"margin,omitempty"` // default: 60
}

// AdvertisingSettings controls ad frequency and placement
type AdvertisingSettings struct {
	BannerType        BannerType        `json:"bannerType,omitempty"`
	DistanceFromMedia UnitsOfMeasure    `json:"distanceFromMedia,omitempty"`
	Frequency         int               `json:"frequency,omitempty"` // 0-10
	Layout            AdvertisingLayout `json:"layout,omitzero"`
}

type AdvertisingLayout struct {
	Margin Margin `json:"margin"`
}

// Metadata contains information about the article
type Metadata struct {
	Authors             []string        `json:"authors,omitempty"`
	CampaignData        json.RawMessage `json:"campaignData,omitempty"`
	CanonicalURL        string          `json:"canonicalURL,omitempty"`
	CoverArt            []CoverArt      `json:"coverArt,omitempty"`
	Excerpt             string          `json:"excerpt,omitempty"`
	Keywords            []string        `json:"keywords,omitempty"`
	ThumbnailURL        string          `json:"thumbnailURL,omitempty"`
	TransparentToolbar  *bool           `json:"transparentToolbar,omitempty"`
	VideoURL            string          `json:"videoURL,omitempty"`
	Links               []LinkedArticle `json:"links,omitempty"`
	DateCreated         *time.Time      `json:"dateCreated,omitempty"`
	DateModified        *time.Time      `json:"dateModified,omitempty"`
	DatePublished       *time.Time      `json:"datePublished,omitempty"`
	GeneratorIdentifier string          `json:"generatorIdentifier,omitempty"`
	GeneratorName       string          `json:"generatorName,omitempty"`
	GeneratorVersion    string          `json:"generatorVersion,omitempty"`
}

// DocumentStyle defines the article's background
type DocumentStyle struct {
	BackgroundColor string `json:"backgroundColor"` // default: "#FFF"
}

// Basic types and enums

// BannerType is an enum for the kind of ad banner
type BannerType string

const (
	BannerTypeAny          BannerType = "any"
	BannerTypeStandard     BannerType = "standard"
	BannerTypeDoubleHeight BannerType = "double_height"
	BannerTypeLarge        BannerType = "large"
)

type FillMode string

const (
	FillModeFit   FillMode = "fit"
	FillModeCover FillMode = "cover"
)

type HorizontalAlignment string

const (
	HorizontalAlignmentLeft   HorizontalAlignment = "left"
	HorizontalAlignmentCenter HorizontalAlignment = "center"
	HorizontalAlignmentRight  HorizontalAlignment = "right"
)

type VerticalAlignment string

const (
	VerticalAlignmentTop    VerticalAlignment = "top"
	VerticalAlignmentCenter VerticalAlignment = "center"
	VerticalAlignmentBottom VerticalAlignment = "bottom"
)

type TextAlignment string

const (
	TextAlignmentLeft      TextAlignment = "left"
	TextAlignmentRight     TextAlignment = "right"
	TextAlignmentCenter    TextAlignment = "center"
	TextAlignmentJustified TextAlignment = "justified"
	TextAlignmentNone      TextAlignment = "none"
)

// UnitsOfMeasure can be an integer or a string with units
type UnitsOfMeasure any

// Margin can be an integer or an object with top/bottom
type Margin any

type MarginObject struct {
	Top    *UnitsOfMeasure `json:"top,omitempty"`
	Bottom *UnitsOfMeasure `json:"bottom,omitempty"`
}

// Padding can be an integer or an object with individual sides
type Padding any

type PaddingObject struct {
	Top    *UnitsOfMeasure `json:"top,omitempty"`
	Right  *UnitsOfMeasure `json:"right,omitempty"`
	Bottom *UnitsOfMeasure `json:"bottom,omitempty"`
	Left   *UnitsOfMeasure `json:"left,omitempty"`
}

// Fill types
type Fill interface {
	fillType() string
}

type ImageFill struct {
	Type                string               `json:"type"` // "image"
	URL                 string               `json:"URL"`
	Attachment          string               `json:"attachment,omitempty"` // "scroll" or "fixed"
	FillMode            *FillMode            `json:"fillMode,omitempty"`
	HorizontalAlignment *HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	VerticalAlignment   *VerticalAlignment   `json:"verticalAlignment,omitempty"`
}

func (ImageFill) fillType() string { return "image" }

type VideoFill struct {
	Type                string               `json:"type"` // "video"
	URL                 string               `json:"URL"`
	StillURL            string               `json:"stillURL"`
	FillMode            *FillMode            `json:"fillMode,omitempty"`
	HorizontalAlignment *HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	VerticalAlignment   *VerticalAlignment   `json:"verticalAlignment,omitempty"`
	Loop                *bool                `json:"loop,omitempty"` // default: true
}

func (VideoFill) fillType() string { return "video" }

type LinearGradient struct {
	Type       string      `json:"type"` // "linear_gradient"
	ColorStops []ColorStop `json:"colorStops"`
	Angle      *float64    `json:"angle,omitempty"`
}

func (LinearGradient) fillType() string { return "linear_gradient" }

type ColorStop struct {
	Color    string   `json:"color"`
	Location *float64 `json:"location,omitempty"` // percentage 0-1
}

// Style types

type ComponentStyle struct {
	BackgroundColor string      `json:"backgroundColor,omitempty"`
	Fill            Fill        `json:"fill,omitempty"`
	Opacity         *float64    `json:"opacity,omitempty"` // percentage 0-1
	Border          *Border     `json:"border,omitempty"`
	TableStyle      *TableStyle `json:"tableStyle,omitempty"`
}

type Border struct {
	All    *StrokeStyle `json:"all,omitempty"`
	Top    *bool        `json:"top,omitempty"`
	Right  *bool        `json:"right,omitempty"`
	Bottom *bool        `json:"bottom,omitempty"`
	Left   *bool        `json:"left,omitempty"`
}

type StrokeStyle struct {
	Color string          `json:"color,omitempty"`
	Width *UnitsOfMeasure `json:"width,omitempty"`
	Style string          `json:"style,omitempty"` // "solid", "dashed", "dotted"
}

// Layout types

type ComponentLayout struct {
	ColumnSpan                 *int                 `json:"columnSpan,omitempty"`
	ColumnStart                *int                 `json:"columnStart,omitempty"`
	ContentInset               any                  `json:"contentInset,omitempty"` // bool or ContentInsetObject
	HorizontalContentAlignment *HorizontalAlignment `json:"horizontalContentAlignment,omitempty"`
	Margin                     Margin               `json:"margin,omitempty"`
	MaximumContentWidth        *UnitsOfMeasure      `json:"maximumContentWidth,omitempty"`
	MinimumHeight              *UnitsOfMeasure      `json:"minimumHeight,omitempty"`
	IgnoreDocumentGutter       any                  `json:"ignoreDocumentGutter,omitempty"` // bool or string
	IgnoreDocumentMargin       any                  `json:"ignoreDocumentMargin,omitempty"` // bool or string
}

type ContentInsetObject struct {
	Top    *bool `json:"top,omitempty"`
	Right  *bool `json:"right,omitempty"`
	Bottom *bool `json:"bottom,omitempty"`
	Left   *bool `json:"left,omitempty"`
}

// Text styles

type TextStyle struct {
	BackgroundColor    string          `json:"backgroundColor,omitempty"`
	FontFamily         string          `json:"fontFamily,omitempty"`
	FontName           string          `json:"fontName,omitempty"`
	FontSize           float64         `json:"fontSize,omitempty"`
	FontStyle          string          `json:"fontStyle,omitempty"`  // "normal", "italic", "oblique"
	FontWeight         any             `json:"fontWeight,omitempty"` // int (100-900) or string
	FontWidth          string          `json:"fontWidth,omitempty"`
	OrderedListItems   *ListItemStyle  `json:"orderedListItems,omitempty"`
	UnorderedListItems *ListItemStyle  `json:"unorderedListItems,omitempty"`
	Strikethrough      any             `json:"strikethrough,omitempty"` // bool or TextDecoration
	Stroke             *TextDecoration `json:"stroke,omitempty"`
	TextColor          string          `json:"textColor,omitempty"`
	TextShadow         *TextShadow     `json:"textShadow,omitempty"`
	Tracking           float64         `json:"tracking,omitempty"`          // percentage
	Underline          any             `json:"underline,omitempty"`         // bool or TextDecoration
	VerticalAlignment  string          `json:"verticalAlignment,omitempty"` // "superscript", "subscript", "baseline"
}

type ComponentTextStyle struct {
	TextStyle
	DropCapStyle           *DropCapStyle  `json:"dropCapStyle,omitempty"`
	FirstLineIndent        *int           `json:"firstLineIndent,omitempty"`
	HangingPunctuation     *bool          `json:"hangingPunctuation,omitempty"`
	Hyphenation            *bool          `json:"hyphenation,omitempty"`
	LineHeight             *int           `json:"lineHeight,omitempty"`
	LinkStyle              *TextStyle     `json:"linkStyle,omitempty"`
	ParagraphSpacingBefore *int           `json:"paragraphSpacingBefore,omitempty"`
	ParagraphSpacingAfter  *int           `json:"paragraphSpacingAfter,omitempty"`
	TextAlignment          *TextAlignment `json:"textAlignment,omitempty"`
}

type DropCapStyle struct {
	NumberOfLines       int    `json:"numberOfLines"` // 2-10
	BackgroundColor     string `json:"backgroundColor,omitempty"`
	FontName            string `json:"fontName,omitempty"`
	NumberOfCharacters  *int   `json:"numberOfCharacters,omitempty"` // 1-4
	NumberOfRaisedLines *int   `json:"numberOfRaisedLines,omitempty"`
	Padding             *int   `json:"padding,omitempty"`
	TextColor           string `json:"textColor,omitempty"`
}

type TextDecoration struct {
	Color string `json:"color,omitempty"`
	Width *int   `json:"width,omitempty"`
}

type TextShadow struct {
	Color   string       `json:"color"`
	Radius  float64      `json:"radius"`            // 0-100
	Opacity *float64     `json:"opacity,omitempty"` // percentage
	Offset  ShadowOffset `json:"offset"`
}

type ShadowOffset struct {
	X float64 `json:"x"` // -50 to 50
	Y float64 `json:"y"` // -50 to 50
}

type ListItemStyle struct {
	Type      string `json:"type"`                // "bullet", "decimal", etc.
	Character string `json:"character,omitempty"` // max 1 char
}

type InlineTextStyle struct {
	RangeLength int `json:"rangeLength"`
	RangeStart  int `json:"rangeStart"`
	TextStyle   any `json:"textStyle"` // string reference or TextStyle object
}

type Component interface {
	componentRole() string
}

type TextComponent struct {
	Role                 string            `json:"role"`
	Anchor               *Anchor           `json:"anchor,omitempty"`
	Identifier           string            `json:"identifier,omitempty"`
	Layout               any               `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any               `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation        `json:"animation,omitempty"`
	Behavior             *Behavior         `json:"behavior,omitempty"`
	Text                 string            `json:"text"`
	Additions            []Addition        `json:"additions,omitempty"`
	Format               string            `json:"format,omitempty"` // "html", "markdown", "none"
	InlineTextStyles     []InlineTextStyle `json:"inlineTextStyles,omitempty"`
	TextStyle            any               `json:"textStyle,omitempty"` // string reference or ComponentTextStyle
	AccessibilityCaption string            `json:"accessibilityCaption,omitempty"`
	Caption              any               `json:"caption,omitempty"` // string or CaptionDescriptor
}

func (TextComponent) componentRole() string { return "text" }

type ImageComponent struct {
	Role                 string     `json:"role"`
	Anchor               *Anchor    `json:"anchor,omitempty"`
	Identifier           string     `json:"identifier,omitempty"`
	Layout               any        `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any        `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation `json:"animation,omitempty"`
	Behavior             *Behavior  `json:"behavior,omitempty"`
	URL                  string     `json:"URL"`
	AccessibilityCaption string     `json:"accessibilityCaption,omitempty"`
	Caption              any        `json:"caption,omitempty"` // string or CaptionDescriptor
}

func (ImageComponent) componentRole() string { return "image" }

type VideoComponent struct {
	Role                 string     `json:"role"`
	Anchor               *Anchor    `json:"anchor,omitempty"`
	Identifier           string     `json:"identifier,omitempty"`
	Layout               any        `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any        `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation `json:"animation,omitempty"`
	Behavior             *Behavior  `json:"behavior,omitempty"`
	URL                  string     `json:"URL"`
	StillURL             string     `json:"stillURL,omitempty"`
	AccessibilityCaption string     `json:"accessibilityCaption,omitempty"`
	Caption              any        `json:"caption,omitempty"`
	AspectRatio          *float64   `json:"aspectRatio,omitempty"` // default: 1.777
	ExplicitContent      *bool      `json:"explicitContent,omitempty"`
}

func (VideoComponent) componentRole() string { return "video" }

type ContainerComponent struct {
	Role           string          `json:"role"`
	Anchor         *Anchor         `json:"anchor,omitempty"`
	Identifier     string          `json:"identifier,omitempty"`
	Layout         any             `json:"layout,omitempty"` // string reference or ComponentLayout
	Style          any             `json:"style,omitempty"`  // string reference or ComponentStyle
	ContentDisplay *ContentDisplay `json:"contentDisplay,omitempty"`
	Components     []Component     `json:"components"`
}

func (ContainerComponent) componentRole() string { return "container" }

type GalleryComponent struct {
	Role       string        `json:"role"`
	Anchor     *Anchor       `json:"anchor,omitempty"`
	Identifier string        `json:"identifier,omitempty"`
	Layout     any           `json:"layout,omitempty"` // string reference or ComponentLayout
	Style      any           `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation  *Animation    `json:"animation,omitempty"`
	Behavior   *Behavior     `json:"behavior,omitempty"`
	Items      []GalleryItem `json:"items"`
}

func (GalleryComponent) componentRole() string { return "gallery" }

type GalleryItem struct {
	URL                  string `json:"URL"`
	AccessibilityCaption string `json:"accessibilityCaption,omitempty"`
	Caption              any    `json:"caption,omitempty"`
	ExplicitContent      *bool  `json:"explicitContent,omitempty"`
}

type DataTableComponent struct {
	Role                 string             `json:"role"`
	Anchor               *Anchor            `json:"anchor,omitempty"`
	Identifier           string             `json:"identifier,omitempty"`
	Layout               any                `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any                `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation         `json:"animation,omitempty"`
	Behavior             *Behavior          `json:"behavior,omitempty"`
	Data                 RecordStore        `json:"data"`
	DataOrientation      string             `json:"dataOrientation,omitempty"` // "horizontal", "vertical"
	ShowDescriptorLabels *bool              `json:"showDescriptorLabels,omitempty"`
	SortBy               []DataTableSorting `json:"sortBy,omitempty"`
}

func (DataTableComponent) componentRole() string { return "datatable" }

type RecordStore struct {
	Descriptors []DataDescriptor `json:"descriptors"`
	Records     []any            `json:"records"` // array of objects
}

type DataDescriptor struct {
	DataType   string      `json:"dataType"` // "string", "text", "image", "number", "integer", "float"
	Format     *DataFormat `json:"format,omitempty"`
	Identifier string      `json:"identifier,omitempty"`
	Key        string      `json:"key"`
	Label      any         `json:"label"` // string or FormattedText
}

type DataFormat interface {
	formatType() string
}

type FloatDataFormat struct {
	Type     string `json:"type"` // "float"
	Decimals *int   `json:"decimals,omitempty"`
}

func (FloatDataFormat) formatType() string { return "float" }

type ImageDataFormat struct {
	Type          string          `json:"type"` // "image"
	MaximumHeight *UnitsOfMeasure `json:"maximumHeight,omitempty"`
	MaximumWidth  *UnitsOfMeasure `json:"maximumWidth,omitempty"`
	MinimumHeight *UnitsOfMeasure `json:"minimumHeight,omitempty"`
	MinimumWidth  *UnitsOfMeasure `json:"minimumWidth,omitempty"`
}

func (ImageDataFormat) formatType() string { return "image" }

type DataTableSorting struct {
	Descriptor string `json:"descriptor"`
	Direction  string `json:"direction"` // "ascending", "descending"
}

type MapComponent struct {
	Role                 string     `json:"role"`
	Anchor               *Anchor    `json:"anchor,omitempty"`
	Identifier           string     `json:"identifier,omitempty"`
	Layout               any        `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any        `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation `json:"animation,omitempty"`
	Behavior             *Behavior  `json:"behavior,omitempty"`
	Latitude             *float64   `json:"latitude,omitempty"`
	Longitude            *float64   `json:"longitude,omitempty"`
	MapType              string     `json:"mapType,omitempty"` // "standard", "hybrid", "satellite"
	Span                 *MapSpan   `json:"span,omitempty"`
	Items                []MapItem  `json:"items,omitempty"`
	AccessibilityCaption string     `json:"accessibilityCaption,omitempty"`
	Caption              any        `json:"caption,omitempty"`
}

func (MapComponent) componentRole() string { return "map" }

type MapSpan struct {
	LatitudeDelta  float64 `json:"latitudeDelta"`  // 0.0-90.0
	LongitudeDelta float64 `json:"longitudeDelta"` // 0.0-180.0
}

type MapItem struct {
	Latitude  float64 `json:"latitude"`  // -90.0 to 90.0
	Longitude float64 `json:"longitude"` // -180.0 to 180.0
	Caption   string  `json:"caption,omitempty"`
}

// Supporting types

type Anchor struct {
	TargetAnchorPosition      VerticalAlignment  `json:"targetAnchorPosition"`
	OriginAnchorPosition      *VerticalAlignment `json:"originAnchorPosition,omitempty"`
	RangeLength               *int               `json:"rangeLength,omitempty"`
	RangeStart                *int               `json:"rangeStart,omitempty"`
	TargetComponentIdentifier string             `json:"targetComponentIdentifier,omitempty"`
	Target                    string             `json:"target,omitempty"`
}

type Animation interface {
	animationType() string
}

type AppearAnimation struct {
	Type             string `json:"type"` // "appear"
	UserControllable *bool  `json:"userControllable,omitempty"`
}

func (AppearAnimation) animationType() string { return "appear" }

type FadeInAnimation struct {
	Type             string   `json:"type"` // "fade_in"
	UserControllable *bool    `json:"userControllable,omitempty"`
	InitialAlpha     *float64 `json:"initialAlpha,omitempty"`
}

func (FadeInAnimation) animationType() string { return "fade_in" }

type Behavior interface {
	behaviorType() string
}

type ParallaxBehavior struct {
	Type   string   `json:"type"`             // "parallax"
	Factor *float64 `json:"factor,omitempty"` // 0.5-2.0
}

func (ParallaxBehavior) behaviorType() string { return "parallax" }

type Addition interface {
	additionType() string
}

type LinkAddition struct {
	Type        string `json:"type"` // "link"
	URL         string `json:"URL"`
	RangeLength int    `json:"rangeLength"`
	RangeStart  int    `json:"rangeStart"`
}

func (LinkAddition) additionType() string { return "link" }

type CaptionDescriptor struct {
	Text             string            `json:"text"`
	Additions        []Addition        `json:"additions,omitempty"`
	Format           string            `json:"format,omitempty"`
	InlineTextStyles []InlineTextStyle `json:"inlineTextStyles,omitempty"`
	TextStyle        any               `json:"textStyle,omitempty"`
}

type ContentDisplay struct {
	Type           string          `json:"type"`                   // "collection"
	Alignment      string          `json:"alignment,omitempty"`    // "left", "center", "right"
	Distribution   string          `json:"distribution,omitempty"` // "narrow", "wide"
	Gutter         *UnitsOfMeasure `json:"gutter,omitempty"`
	MaximumWidth   *UnitsOfMeasure `json:"maximumWidth,omitempty"`
	MinimumWidth   *UnitsOfMeasure `json:"minimumWidth,omitempty"`
	RowSpacing     *UnitsOfMeasure `json:"rowSpacing,omitempty"`
	VariableSizing *bool           `json:"variableSizing,omitempty"`
	Widows         string          `json:"widows,omitempty"` // "equalize", "optimize"
}

type CoverArt struct {
	Type                 string `json:"type"` // "image"
	URL                  string `json:"URL"`
	AccessibilityCaption string `json:"accessibilityCaption,omitempty"`
}

type LinkedArticle struct {
	URL          string `json:"URL"`
	Relationship string `json:"relationship"` // "related", "promoted"
}

type TableStyle struct {
	Cells         *TableCellStyle   `json:"cells,omitempty"`
	Columns       *TableColumnStyle `json:"columns,omitempty"`
	HeaderCells   *TableCellStyle   `json:"headerCells,omitempty"`
	HeaderColumns *TableColumnStyle `json:"headerColumns,omitempty"`
	HeaderRows    *TableRowStyle    `json:"headerRows,omitempty"`
	Rows          *TableRowStyle    `json:"rows,omitempty"`
}

type TableCellStyle struct {
	BackgroundColor     string                      `json:"backgroundColor,omitempty"`
	Border              *Border                     `json:"border,omitempty"`
	Width               *int                        `json:"width,omitempty"`
	MinimumWidth        *UnitsOfMeasure             `json:"minimumWidth,omitempty"`
	Height              *UnitsOfMeasure             `json:"height,omitempty"`
	Padding             Padding                     `json:"padding,omitempty"`
	TextStyle           any                         `json:"textStyle,omitempty"`
	HorizontalAlignment *HorizontalAlignment        `json:"horizontalAlignment,omitempty"`
	VerticalAlignment   *VerticalAlignment          `json:"verticalAlignment,omitempty"`
	Conditional         []ConditionalTableCellStyle `json:"conditional,omitempty"`
}

type ConditionalTableCellStyle struct {
	TableCellStyle
	Selectors []TableCellSelector `json:"selectors"`
}

type TableCellSelector struct {
	Descriptor  string `json:"descriptor,omitempty"`
	ColumnIndex *int   `json:"columnIndex,omitempty"`
	RowIndex    *int   `json:"rowIndex,omitempty"`
	EvenColumns *bool  `json:"evenColumns,omitempty"`
	EvenRows    *bool  `json:"evenRows,omitempty"`
	OddColumns  *bool  `json:"oddColumns,omitempty"`
	OddRows     *bool  `json:"oddRows,omitempty"`
}

type TableColumnStyle struct {
	BackgroundColor string                        `json:"backgroundColor,omitempty"`
	Width           *int                          `json:"width,omitempty"`
	MinimumWidth    *UnitsOfMeasure               `json:"minimumWidth,omitempty"`
	Divider         *StrokeStyle                  `json:"divider,omitempty"`
	Conditional     []ConditionalTableColumnStyle `json:"conditional,omitempty"`
}

type ConditionalTableColumnStyle struct {
	TableColumnStyle
	Selectors []TableColumnSelector `json:"selectors"`
}

type TableColumnSelector struct {
	Descriptor  string `json:"descriptor,omitempty"`
	ColumnIndex *int   `json:"columnIndex,omitempty"`
	Odd         *bool  `json:"odd,omitempty"`
	Even        *bool  `json:"even,omitempty"`
}

type TableRowStyle struct {
	BackgroundColor string                     `json:"backgroundColor,omitempty"`
	Height          *int                       `json:"height,omitempty"`
	Divider         *StrokeStyle               `json:"divider,omitempty"`
	Conditional     []ConditionalTableRowStyle `json:"conditional,omitempty"`
}

type ConditionalTableRowStyle struct {
	TableRowStyle
	Selectors []TableRowSelector `json:"selectors"`
}

type TableRowSelector struct {
	Descriptor string `json:"descriptor,omitempty"`
	RowIndex   *int   `json:"rowIndex,omitempty"`
	Odd        *bool  `json:"odd,omitempty"`
	Even       *bool  `json:"even,omitempty"`
}

type FormattedText struct {
	Additions        []Addition        `json:"additions,omitempty"`
	Format           string            `json:"format,omitempty"` // "html", "none"
	InlineTextStyles []InlineTextStyle `json:"inlineTextStyles,omitempty"`
	Text             string            `json:"text"`
	TextStyle        any               `json:"textStyle,omitempty"`
	Type             string            `json:"type"` // "formatted_text"
}

type DividerComponent struct {
	Role       string       `json:"role"`
	Anchor     *Anchor      `json:"anchor,omitempty"`
	Identifier string       `json:"identifier,omitempty"`
	Layout     any          `json:"layout,omitempty"` // string reference or ComponentLayout
	Style      any          `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation  *Animation   `json:"animation,omitempty"`
	Behavior   *Behavior    `json:"behavior,omitempty"`
	Stroke     *StrokeStyle `json:"stroke,omitempty"`
}

func (DividerComponent) componentRole() string { return "divider" }

type PhotoComponent struct {
	Role                 string     `json:"role"`
	Anchor               *Anchor    `json:"anchor,omitempty"`
	Identifier           string     `json:"identifier,omitempty"`
	Layout               any        `json:"layout,omitempty"` // string reference or ComponentLayout
	Style                any        `json:"style,omitempty"`  // string reference or ComponentStyle
	Animation            *Animation `json:"animation,omitempty"`
	Behavior             *Behavior  `json:"behavior,omitempty"`
	URL                  string     `json:"URL"`
	AccessibilityCaption *string    `json:"accessibilityCaption,omitempty"`
	Caption              any        `json:"caption,omitempty"` // string or CaptionDescriptor
}

func (PhotoComponent) componentRole() string { return "photo" }

type Response struct {
	Data Data `json:"data"`
	Meta Meta `json:"meta"`
}

type Data struct {
	CreatedAt                   time.Time `json:"createdAt"`
	ModifiedAt                  time.Time `json:"modifiedAt"`
	ID                          string    `json:"id"`
	Type                        string    `json:"type"`
	ShareURL                    string    `json:"shareUrl"`
	Links                       Links     `json:"links"`
	Document                    Article   `json:"document"`
	Revision                    string    `json:"revision"`
	State                       string    `json:"state"`
	AccessoryText               any       `json:"accessoryText"`
	Title                       string    `json:"title"`
	MaturityRating              any       `json:"maturityRating"`
	Warnings                    any       `json:"warnings"`
	TargetTerritoryCountryCodes []string  `json:"targetTerritoryCountryCodes"`
	IsCandidateToBeFeatured     bool      `json:"isCandidateToBeFeatured"`
	IsSponsored                 bool      `json:"isSponsored"`
	IsPreview                   bool      `json:"isPreview"`
	IsDevelopingStory           bool      `json:"isDevelopingStory"`
	IsHidden                    bool      `json:"isHidden"`
}

type Links struct {
	Channel  string   `json:"channel"`
	Self     string   `json:"self"`
	Sections []string `json:"sections"`
}

type Meta struct {
	Throttling Throttling `json:"throttling"`
}

type Throttling struct {
	IsThrottled             bool `json:"isThrottled"`
	QueueSize               int  `json:"queueSize"`
	EstimatedDelayInSeconds int  `json:"estimatedDelayInSeconds"`
	QuotaAvailable          int  `json:"quotaAvailable"`
}
