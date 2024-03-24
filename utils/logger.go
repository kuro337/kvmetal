package utils

import (
	"fmt"
	"log"
	"strings"
)

type Styles int

const (
	Bold Styles = 1 << iota
	Italic
	Dim
	Underline
	NoUnderline
	Reset
)

type ANSISpecial string

const (
	TickGreen ANSISpecial = "\x1b[1m\x1b[32m✓\x1b[0m"
	CrossRed  ANSISpecial = "\x1b[1m\x1b[31m✗\x1b[0m"
)

type Delimiter string

const (
	DelimiterEq    Delimiter = "======================================================================"
	DelimiterEqDim           = "\033[2m======================================================================\033[0m"
	DelimiterStar            = "**********************************************************************"
	DelimiterDash            = "-----------------------------------------------------"
)

type Color string

const (
	BlueHC        Color = "\033[94m"
	GreenHC             = "\033[92m"
	PurpHC              = "\033[95m"
	LightGrey           = "\033[37m"
	Red                 = "\033[0;31m"
	Green               = "\033[0;32m"
	Blue                = "\033[0;34m"
	Yellow              = "\033[0;33m"
	LightYellow         = "\033[1;33m"
	PurpleDull          = "\x1b[38;5;5m"
	GreenDeepdark       = "\x1b[38;5;36m"
	GreenLight          = "\x1b[38;5;49m"
	SkyBlue             = "\x1b[38;5;45m"
	CoolBlue            = "\x1b[38;5;44m"
	Greenish            = "\x1b[38;5;42m"
	PurpCool            = "\x1b[38;5;93m"
	Indigo              = "\x1b[38;5;57m"
	BlueDeep            = "\x1b[38;5;39m"
	BlueHyper           = "\x1b[38;5;33m"
	GreenWhite          = "\x1b[38;5;121m"
	WhitePink           = "\x1b[38;5;218m"
	Sand                = "\x1b[38;5;222m"
	Peach               = "\x1b[38;5;230m"
	Pinkish             = "\x1b[38;5;199m"
	WhiteBlue           = "\x1b[38;5;117m"
	PurpDark            = "\x1b[38;5;128m"
	PurpShine           = "\x1b[38;5;129m"
	PurpWhite           = "\x1b[38;5;189m"
	GreyBlue            = "\x1b[38;5;153m"
	GreenGrey           = "\x1b[38;5;158m"
	LightPink           = "\x1b[38;5;219m"
	Purple              = "\x1b[38;5;135m"
)

/*
Create a Formatter to Format Strings
*/
type LogFormat struct {
	Color      Color
	KeyColor   Color
	ValColor   Color
	color      string
	keyColor   string
	valColor   string
	StyleFlags Styles
	KeyOnly    bool
	ValueOnly  bool
}

type FormatOps interface{}

// Formatter defines an interface for formatting strings.
type Formatter interface {
	Format(string) string
	FormatKV(...string) string
	Update(interface{})
}

type formatter struct {
	config LogFormat
}

func (f *formatter) setColors() {
	f.config.color = string(f.config.Color)
	f.config.keyColor = string(f.config.KeyColor)
	f.config.valColor = string(f.config.ValColor)
}

/*
Create a Formatter to Format Strings

	config := utils.LogFormat{StyleFlags: utils.Bold}
	f := utils.NewFormatter(config)

	str := f.Format("This text will be bold")
	t.Log(str)

	// Format with Bold , Italics , and Color
	f.Update(utils.LogFormat{Color: utils.CoolBlue, StyleFlags: utils.Bold | utils.Italic})
	boldItalicColored := f.FormatKV("Key1", "Value Only Color", "Key2", "Value Only Color")
	t.Log(boldItalicColored)

	f.Update(utils.LogFormat{StyleFlags: utils.Bold, KeyOnly: true})
	formattedKV := f.FormatKV("Keys Bold", "Value Not Bold", "Key2 Bold", "Not Bold")

	t.Log(formattedKV)

	f.Update(utils.LogFormat{ValColor: utils.PurpHC, ValueOnly: true})
	valColored := f.FormatKV("Key1", "Value Only Color", "Key2", "Value Only Color")
	t.Log(valColored)
*/
func NewFormatter(config LogFormat) Formatter {
	config.color = string(config.Color)
	config.keyColor = string(config.KeyColor)
	config.valColor = string(config.ValColor)

	return &formatter{config: config}
}

/*
Update with a new LogFormat Configuration
*/
func (f *formatter) Update(config interface{}) {
	if cfg, ok := config.(LogFormat); ok {
		f.config = cfg
		f.setColors()
	} else {
		log.Print("UpdateConfig: provided configuration is not of type LogFormat")
	}
}

func (opts *formatter) Format(s string) string {
	var sb strings.Builder

	// Apply color
	if opts.config.Color != "" {
		sb.WriteString(opts.config.color)
	}

	/* Apply Styles for Bold, Italics, Underline...*/
	if opts.config.StyleFlags&Bold != 0 {
		sb.WriteString(BOLD)
	}
	if opts.config.StyleFlags&Italic != 0 {
		sb.WriteString(ITALIC) // Similar assumption for Italic
	}
	if opts.config.StyleFlags&Underline != 0 {
		sb.WriteString(UNDERLINE) // And so on for Underline
	}

	if opts.config.Color != "" {
		sb.WriteString(opts.config.color)
	}
	// Append the string
	sb.WriteString(s)

	// Reset formatting at the end
	sb.WriteString(NC)

	return sb.String()
}

func (opts *formatter) applyStyles(s string, isKey bool) string {
	var sb strings.Builder

	// Apply styles based on the context (key or value)
	if (!opts.config.KeyOnly && !opts.config.ValueOnly) || (opts.config.KeyOnly && isKey) || (opts.config.ValueOnly && !isKey) {
		if opts.config.StyleFlags&Bold != 0 {
			sb.WriteString(BOLD)
		}
		if opts.config.StyleFlags&Italic != 0 {
			sb.WriteString(ITALIC)
		}
		if opts.config.StyleFlags&Underline != 0 {
			sb.WriteString(UNDERLINE)
		}

		// Apply color based on the context (key or value)
		color := opts.config.color // Default color
		if isKey && opts.config.KeyColor != "" {
			color = opts.config.keyColor
		} else if !isKey && opts.config.ValColor != "" {
			color = opts.config.valColor
		}

		sb.WriteString(color)
	}

	// Append the string and reset formatting
	sb.WriteString(s)
	sb.WriteString(NC)
	return sb.String()
}

/*
formatter := utils.LogFormat{Bold: true, Color: utils.GREEN_DEEPDARK}

log.Printf(formatter.FormatKV("Key1", "Value1", "Key2", "Value2"))
*/
func (opts *formatter) FormatKV(kvs ...string) string {
	var sb strings.Builder

	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i]
		var val string
		if i+1 < len(kvs) {
			val = kvs[i+1]
		}

		formattedKey := opts.applyStyles(key, true)
		formattedVal := opts.applyStyles(val, false)

		// Append formatted key and value
		sb.WriteString(formattedKey)
		sb.WriteString(": ")
		sb.WriteString(formattedVal)

		if i+2 < len(kvs) { // Add newline if not the last pair
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

const (
	NC             = "\033[0m"
	BOLD           = "\033[1m"
	DIM            = "\033[2m"
	ITALIC         = "\033[3m"
	UNDERLINE      = "\033[4m"
	NO_UNDERLINE   = "\033[24m"
	BLUE_HI        = "\033[94m"
	GREEN_HI       = "\033[92m"
	PURP_HI        = "\033[95m"
	LIGHT_GREY     = "\033[37m"
	RED            = "\033[0;31m"
	GREEN          = "\033[0;32m"
	BLUE           = "\033[0;34m"
	YELLOW         = "\033[0;33m"
	LIGHT_YELLOW   = "\033[1;33m"
	PURPLE_DULL    = "\x1b[38;5;5m"
	GREEN_DEEPDARK = "\x1b[38;5;36m"
	GREEN_LIGHT    = "\x1b[38;5;49m"
	SKYBLUE        = "\x1b[38;5;45m"
	COOLBLUE       = "\x1b[38;5;44m"
	GREENISH       = "\x1b[38;5;42m"
	PURPCOOL       = "\x1b[38;5;93m"
	INDIGO         = "\x1b[38;5;57m"
	BLUE_DEEP      = "\x1b[38;5;39m"
	BLUEHYPER      = "\x1b[38;5;33m"
	GREEN_WHITE    = "\x1b[38;5;121m"
	WHITEPINK      = "\x1b[38;5;218m"
	SAND           = "\x1b[38;5;222m"
	PEACH          = "\x1b[38;5;230m"
	PINKISH        = "\x1b[38;5;199m"
	WHITEBLUE      = "\x1b[38;5;117m"
	PURP_DARK      = "\x1b[38;5;128m"
	PURP_SHINE     = "\x1b[38;5;129m"
	PURPLE_WHITE   = "\x1b[38;5;189m"
	GREY_BLUE      = "\x1b[38;5;153m"
	GREEN_GREY     = "\x1b[38;5;158m"
	LIGHTPINK      = "\x1b[38;5;219m"
	PURPLE         = "\x1b[38;5;135m"
	TICK_GREEN     = "\x1b[1m\x1b[32m✓\x1b[0m"
	CROSS_RED      = "\x1b[1m\x1b[31m✗\x1b[0m"

	DELIMITER    = "======================================================================"
	DELIMITERDIM = "\033[2m======================================================================\033[0m"
	SECTION      = "**********************************************************************"
	DOTTED       = "-----------------------------------------------------"
)

// FormatOptions contains formatting options for a key-value pair.
type FormatOptions struct {
	Bold  bool
	Color string
}

/*
FormatKV formats a series of key-value pairs according to the given format options.

	Usage:

	fmt.Println(FormatKV(FormatOptions{Bold: true, Color: RED}, "Key1", "Value1", "Key2", "Value2"))
	fmt.Println(FormatKV(FormatOptions{Bold: true}, "Key3", "Value3", "Key4", "Value4", "Key5", "Value5"))
	fmt.Println(FormatKV(FormatOptions{Color: RED}, "Key6", "Value6"))
*/
func FormatKV(opts FormatOptions, kvs ...string) string {
	var sb strings.Builder

	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i]
		var val string
		if i+1 < len(kvs) {
			val = kvs[i+1]
		}

		// Start formatting
		if opts.Bold {
			sb.WriteString(BOLD)
		}

		if opts.Color != "" {
			sb.WriteString(opts.Color)
		}

		// Append key
		sb.WriteString(key)
		sb.WriteString(": ")

		// Append value (keep applying color and bold)
		sb.WriteString(val)

		// Reset formatting at the end of each key-value pair
		sb.WriteString(NC)

		if i+2 < len(kvs) { // Add newline if not the last pair
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

/*
Prints a Formatted Result Block with a Title and K/V pairs of Header and []string

Usage:

	func main() {
		stopsArr := []string{"Stop 1", "Stop 2"}
		startsArr := []string{"Start 1", "Start 2"}

		result := GetResultBlock("Results of Processing Something", "Stops", stopsArr, "Starts", startsArr)
		fmt.Println(result)
	}
*/
func GetResultBlock(title string, arrays ...interface{}) string {
	var builder strings.Builder
	mainTitle := TurnContentBoldColorDelimited(title, PURP_HI, PURP_DARK, 5)
	builder.WriteString(mainTitle)

	for i := 0; i < len(arrays); i += 2 {
		arrayTitle, ok := arrays[i].(string)
		if !ok || i+1 >= len(arrays) {
			continue
		}
		arrayValues, ok := arrays[i+1].([]string)
		if !ok {
			continue // Skip if the next item is not an array of strings
		}

		// Add array title
		builder.WriteString(TurnBoldBlueDelimited(arrayTitle))

		// Add array values
		for _, value := range arrayValues {
			builder.WriteString(fmt.Sprintf("%s\n", value))
		}
	}

	return builder.String()
}

func TurnKeyBold(key, val string) string {
	return fmt.Sprintf("%s%s%s%s", BOLD, key, NC, val)
}

func GetDelimiter(color string, highlighted, dim bool) string {
	delimiter := DELIMITER

	if dim {
		delimiter = DELIMITERDIM
	}
	str := fmt.Sprintf("%s%s%s", color, delimiter, NC)

	if highlighted {
		return fmt.Sprintf("%s%s", BOLD, str)
	}
	return str
}

/*
TurnContentBoldColorDelimited returns a formatted string with bold color and delimiters, including leading spaces.
Usage:

	mainTitle := TurnContentBoldColorDelimited(title, PURP_HI, PURP_DARK, 5)
*/
func TurnContentBoldColorDelimited(msg, msgColor, delimiterColor string, spacing int) string {
	return fmt.Sprintf("%s%s%s\n%s%*s%s%s%s\n%s%s%s%s\n", BOLD, delimiterColor, DELIMITERDIM, NC, spacing, BOLD, msgColor, msg, NC, BOLD, delimiterColor, DELIMITERDIM, NC)
}

/*
Returns a String with the msg , a newline, and a delimiter, then a newline
Usage:
*/
func TurnColorWithNewlineDelimiter(msg, color string) string {
	return fmt.Sprintf("%s%s%s\n%s\n", BOLD, color, msg, DOTTED)
}

/*
Structures the Result with a Colored Heading and surrounds it with Delimiters

		log.Print(utils.StructureResultWithHeadingAndColoredMsg(
		"CloudInit UserData Set To", utils.PEACH,
		userDataContent,
	))
*/
func StructureResultWithHeadingAndColoredMsg(msg, color, content string) string {
	return fmt.Sprintf("%s%s%s%s\n%s\n%s\n%s\n", BOLD, color, msg, NC, DOTTED, content, DOTTED)
}

/*
Make the Second Part of a String Colored and Bold

	log.Print(utils.TurnValBoldColor("Preset: ", preset, utils.PURP_HI))
	// Preset: redpanda
*/
func TurnValBoldColor(key, val, color string) string {
	return fmt.Sprintf("%s%s%s%s%s", key, BOLD, color, val, NC)
}

func TurnError(msg string) string {
	return fmt.Sprintf("%s%s%s", RED, msg, NC)
}

func TurnSuccess(msg string) string {
	return fmt.Sprintf("%s%s%s", GREEN_HI, msg, NC)
}

func TurnWarning(msg string) string {
	return fmt.Sprintf("%s%s%s%s", DIM, YELLOW, msg, NC)
}

func TurnValBold(key, val string) string {
	return fmt.Sprintf("%s%s%s%s", key, BOLD, val, NC)
}

func TurnKeyColor(key, val, color string) string {
	return fmt.Sprintf("%s%s%s%s%s", BOLD, color, key, NC, val)
}

func TurnValColor(key, val, color string) string {
	return fmt.Sprintf("%s%s%s%s%s", key, BOLD, color, val, NC)
}

func TurnKeyBoldColor(key, val, color string) string {
	return fmt.Sprintf("%s%s%s%s%s", BOLD, color, key, NC, val)
}

func TurnBlueDelimited(msg string) string {
	return fmt.Sprintf("%s%s\n%s%s\n%s%s%s\n", BLUE_DEEP, DELIMITERDIM, NC, msg, BLUE_DEEP, DELIMITERDIM, NC)
}

func TurnBoldColorDelimited(msg, color string) string {
	return fmt.Sprintf("%s%s%s\n%s%s%s\n%s%s%s\n", BOLD, color, DELIMITERDIM, NC, BOLD, msg, color, DELIMITERDIM, NC)
}

func TurnBoldBlueDelimited(msg string) string {
	return fmt.Sprintf("%s%s%s\n%s%s%s\n%s%s%s\n", BOLD, BLUE_DEEP, DELIMITERDIM, NC, BOLD, msg, BLUE_DEEP, DELIMITERDIM, NC)
}

func TurnColorDelimited(msg, color string) string {
	return fmt.Sprintf("%s%s\n%s%s\n%s%s%s\n", color, DELIMITERDIM, NC, msg, color, DELIMITERDIM, NC)
}

func AddDelimiter(msg string) string {
	return fmt.Sprintf("%s\n%s\n%s", DELIMITERDIM, msg, DELIMITERDIM)
}

func TurnBold(msg string) string { return fmt.Sprintf("%s%s%s\n", BOLD, msg, NC) }

func TurnColorBold(msg, color string) string {
	return fmt.Sprintf("%s%s%s%s%s\n", BOLD, color, msg, color, NC)
}

func TurnUnderline(msg string) string { return fmt.Sprintf("%s%s%s\n", UNDERLINE, msg, NC) }

func TurnBoldColor(msg, color string) string { return fmt.Sprintf("%s%s%s%s\n", BOLD, color, msg, NC) }

func TurnColor(msg, color string) string { return fmt.Sprintf("%s%s%s\n", color, msg, NC) }

func TurnSand(msg string) string { return fmt.Sprintf("%s%s%s\n", SAND, msg, NC) }

// Enables Logging Flags that display Date, Timestamp, Caller File, and Caller Line Number
func EnableInfo() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func DisableInfo() {
	log.SetFlags(0)
}

/*
		// Example with color
		coloredMsg := LogFormat("This section is in purple", Purple)
		fmt.Println(coloredMsg)

		// Example with color and style
		coloredStyledMsg := LogFormat("This section is light pink and bold", LightPink, Bold)
		fmt.Println(coloredStyledMsg)
	}
*/
func FormatLog(msg string, styles ...string) string {
	startFormat := ""
	for _, style := range styles {
		startFormat += style
	}

	return fmt.Sprintf("%s%s%s", startFormat, msg, NC)
}

func LogKVResults(keyColor, valueColor string, droneStats map[string]string) {
	fmt.Println(DELIMITERDIM)
	log.Printf("%s%sDrone Stats%s", BOLD, UNDERLINE, NC)

	for key, value := range droneStats {
		log.Printf("%s%-20s: %s%s%s\n", keyColor, key, valueColor, value, NC)
	}

	log.Println(DELIMITERDIM)
}

func LogError(msg string) {
	log.Printf("%s%s%s", RED, msg, NC)
}

func LogWarning(msg string) {
	log.Printf("%s%s%s%s", DIM, YELLOW, msg, NC)
}

func LogMainAction(msg string) string {
	return fmt.Sprintf("%s\n%s%s%s%s\n%s\n", DELIMITER, BOLD, WHITEBLUE, msg, NC, DELIMITER)
}

func LogDottedLineDelimitedText(msg string) {
	fmt.Printf("%s\n%s\n%s\n", DOTTED, msg, DOTTED)
}

func LogSection(msg string) string {
	return fmt.Sprintf("%s%s%s\n%s%s%s%s\n%s%s%s\n", DIM, DELIMITER, NC, BOLD, GREEN_HI, msg, NC, DIM, DELIMITER, NC)
}

func LogStep(msg string) {
	fmt.Printf("%s%s%s\n%s%s%s%s\n%s%s%s\n", DIM, SECTION, NC, BOLD, GREY_BLUE, msg, NC, DIM, SECTION, NC)

	// fmt.Printf("%s\n%s%s%s%s\n%s\n", SECTION, BOLD, GREY_BLUE, msg, NC, SECTION)
}

func LogStepSuccess(msg string) {
	log.Printf("%s%s%s", BOLD, msg, NC)
}

func LogSuccessDark(msg string) {
	log.Printf("%s%s%s", GREEN_DEEPDARK, msg, NC)
}

func LogSuccess(msg string) {
	log.Printf("%s%s%s", GREEN_WHITE, msg, NC)
}

func LogBoldUnderlined(msg string) {
	log.Printf("%s%s%s%s", UNDERLINE, BOLD, msg, NC)
}

func LogBold(msg string) {
	log.Printf("%s%s%s", BOLD, msg, NC)
}

func LogWhiteBlueBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, WHITEBLUE, msg, NC)
}

func LogWhiteBlue(msg string) {
	log.Printf("%s%s%s", WHITEBLUE, msg, NC)
}

func LogGreyBlueBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, GREY_BLUE, msg, NC)
}

func LogGreyBlue(msg string) {
	log.Printf("%s%s%s", GREY_BLUE, msg, NC)
}

func LogRichLightPurpleBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, PURP_HI, msg, NC)
}

func TurnRichLightPurple(msg string) string {
	return fmt.Sprintf("%s%s%s", PURP_HI, msg, NC)
}

func LogTealDark(msg string) {
	log.Printf("%s%s%s", GREEN_DEEPDARK, msg, NC)
}

func LogTealDarkBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, GREEN_DEEPDARK, msg, NC)
}

func LogOffwhite(msg string) {
	log.Printf("%s%s%s%s", BOLD, PEACH, msg, NC)
}

func LogOffwhiteBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, PEACH, msg, NC)
}

func LogSkyBlue(msg string) {
	log.Printf("%s%s%s", BLUE_HI, msg, NC)
}

func LogSkyBlueBold(msg string) {
	log.Printf("%s%s%s%s", BOLD, BLUE_HI, msg, NC)
}

func LogEvent(pairs ...interface{}) {
	message := ""
	for i := 0; i < len(pairs); i += 2 {
		desc := pairs[i]
		var valueStr string
		if i+1 < len(pairs) {

			value := pairs[i+1]
			valueStr = fmt.Sprintf("%s%+v%s", PEACH, value, NC)
		} else {
			valueStr = fmt.Sprint(desc)
		}
		message += fmt.Sprintf("%s%s%s: %s ", BOLD, desc, NC, valueStr)
	}
	log.Print(message)
}

func LogEventCoded(ansicolor string, pairs ...interface{}) {
	message := ""
	for i := 0; i < len(pairs); i += 2 {
		desc := pairs[i]
		var valueStr string
		if i+1 < len(pairs) {

			value := pairs[i+1]
			valueStr = fmt.Sprintf("%s%+v%s", ansicolor, value, NC)
		} else {
			valueStr = fmt.Sprint(desc)
		}
		message += fmt.Sprintf("%s%s%s: %s ", BOLD, desc, NC, valueStr)
	}
	log.Print(message)
}

func Help() {
	fmt.Printf("%s\n%s%s%s%s\n%s\n", DELIMITER, BOLD, WHITEBLUE, " m8l - Kubernetes on Bare Metal using KVM", NC, DELIMITER)

	fmt.Printf(UNDERLINE + "\nSystem Library to Manage virtual machines and launch and configure Kubernetes clusters.\n\n" + NC)

	fmt.Printf("@author " + WHITEBLUE + "kuro337\n\n" + NC)
	fmt.Println(DIM + SECTION + NC + BOLD + PURPLE_WHITE + "\nHost System Linux Deps\n" + NC + DIM + SECTION + NC)

	fmt.Println(`
sudo apt install -y qemu qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager cloud-image-utils libguestfs-tools

sudo reboot
	`)
	fmt.Println(DIM + SECTION + NC + BOLD + PURPLE_DULL + "\nCreating and Managing VMs\n" + NC + DIM + SECTION + NC)

	fmt.Printf(`
%s%sCreating a writable clone of a boot drive%s

qemu-img create -b ubuntu-18.04-server-cloudimg-amd64.img -F qcow2 -f qcow2 ubuntu-vm-disk.qcow2 20G

%s%sStarting a VM with virt-install%s

virt-install --name ubuntu-vm \
	--virt-type kvm \
	--os-type Linux --os-variant ubuntu18.04 \
	--memory 2048 \
	--vcpus 2 \
	--boot hd,menu=on \
	--disk path=ubuntu-vm-disk.qcow2,device=disk \
	--disk path=user-data.img,format=raw \
	--graphics none \
	--noautoconsole
		
	`, BOLD, WHITEPINK, NC, BOLD, WHITEPINK, NC)

	fmt.Println()

	fmt.Println(DIM + SECTION + NC + BOLD + SAND + "\nVirsh Networking and Common Issues\n" + NC + DIM + SECTION + NC)

	fmt.Printf(`
%sIf default network bridge is not seen on normal user%s

virsh net-list --all 	  # default missing 
sudo virsh net-list --all # shows default 

%sMake sure current user is added to the libvirt group%s

%s%sRestarting libvirtd and adding user to libvirt group%s

sudo systemctl restart libvirtd
sudo usermod -aG libvirt <username>
cp /etc/libvirt/libvirt.conf ~/.config/libvirt/
virsh net-autostart default 
sudo reboot

%sUpdate libvirt.conf to set the correct group%s

%s1. Make sure user is set to <user> and group is libvirt%s
# by default it will be root and root

sudo vi /etc/libvirt/qemu.conf

# default 

user = "root"     
group = "root"

# update to current user and group libvirt (recommended)

user = "kuro"     
group = "libvirt"

%s2. Uncomment last line at vi /etc/libvirt/libvirt.conf%s

sudo vi /etc/libvirt/libvirt.conf

#uri_default="qemu:///system" # uncomment this 

%s3. Check if default bridge is available for current user%s

virsh net-list --all
	`, SAND, NC, BOLD, NC, UNDERLINE, PEACH, NC, SAND, NC, BOLD, NC, PEACH, NC, PEACH, NC)

	fmt.Println()
	// Virsh Commands and Usage
	fmt.Println(DIM + SECTION + NC + BOLD + WHITEPINK + "\nVirsh Commands and Usage\n" + NC + DIM + SECTION + NC)
	fmt.Println()
	fmt.Println(LIGHTPINK + "Common virsh VM actions:" + NC)

	fmt.Printf(`
%svirsh commands to interact with Virtual Machines%s
virsh list --all
virsh shutdown ubuntu-vm
virsh suspend ubuntu-vm
virsh resume ubuntu-vm

%sAdd user to libvirt group%s
sudo systemctl restart libvirtd
sudo usermod -aG libvirt kuro
sudo reboot

%s%sCLI util to query VM metadata%s

%ssudo apt-get install arp-scan%s

%sGetting VM MAC & IP Addr%s

virsh dumpxml worker | grep 'mac address'
sudo arp-scan --interface=virbr0 --localnet | grep "52:54:00:25:40:cb"

%slibvirt utils to read write%s

sudo virt-ls -d <vmname> /path/on/vm/
sudo virt-copy-out -d vmname /path/on/vm/init.log /local/
sudo virt-copy-in -a pathto/vm-disk.qcow2 <file_to_copy> /path/in/vm

%sSingle Command to get the IP of a VM from the domain name%s

VM=mydomain sudo arp-scan --interface=virbr0 --localnet \
	| grep -f <(virsh dumpxml $VM \
	| awk -F"'" '/mac address/{print $2}') \
	| awk '{print $1}'

128.999.45.100 %s# ip of VM mydomain%s
`, PURPLE_WHITE, NC, PURPLE_WHITE, NC, BOLD, PURPLE_WHITE, NC, BOLD, NC, PURPLE_WHITE, NC, PURPLE_WHITE, NC, PURPLE_WHITE, NC, BOLD, NC)

	fmt.Println()

	fmt.Println(DIM + SECTION + NC + BOLD + GREEN_LIGHT + "\nLibrary Usage\n" + NC + DIM + SECTION + NC)

	fmt.Printf(`
%sLaunching a Single Control Plane Single Worker Cluster%s

func main() {

	vm.LaunchKubeControlNode()
	vm.LaunchKubeWorkerNode()

	healthy, _ := vm.ClusterHealthCheck()
	if healthy {
		fmt.Println("Cluster is healthy.")
	} 

	vm.FullCleanup("kubecontrol")
	vm.FullCleanup("kubeworker")
}

%sConfiguring a VM according to Custom Specs%s

func main() {
  config := NewVMConfig("kubecontrol").
	SetImageURL("https://cloud-images/ubuntu-22.04.amd64.img"). %s// base Image for VM%s 
	SetBootFilesDir("data/scripts/master_kube"). %s// location to Init Scripts and Systemd Services for VM%s
	DefaultUserData(). %s// Use default user data for username password login%s  
	SetBootServices([]string{"kubemaster.service"}). %s// Define Services to Launch at Startup%s
	SetCores(2).
	SetMemory(2048). %s// Once VM is launched - define the artifacts to pull from the VM%s
	SetArtifacts([]string{"/home/ubuntu/kubeadm-init.log"})

	config.CreateImage().    %s// Create & Cache a Custom Image ,Setup the VM, Launch it%s
		Setup().         %s// Setup the VM%s 
		Launch().        %s// Launch the VM using the KVM Hypervisor%s 
		PullArtifacts(). %s// Pull the Boot Artifacts to Host,%s
		HealthCheck()    %s// Performs a Health Check by testing a deployment and networking%s
}
	`, COOLBLUE, NC, COOLBLUE, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC, GREEN_GREY, NC)

	fmt.Println()
}

func MockANSIPrint() {
	colors := map[string]string{
		"Normal":                NC,
		"Bold":                  BOLD,
		"Dim":                   DIM,
		"Underline":             UNDERLINE,
		"No Underline":          NO_UNDERLINE,
		"Cool Blue":             COOLBLUE,
		"Deep Blue":             BLUE_DEEP,
		"High Intensity Blue":   BLUE_HI,
		"High Intensity Green":  GREEN_HI,
		"High Intensity Purple": PURP_HI,
		"Light Grey":            LIGHT_GREY,
		"Red":                   RED,
		"Green":                 GREEN,
		"Blue":                  BLUE,
		"Yellow":                YELLOW,
		"Light Yellow":          LIGHT_YELLOW,
		"Purple Dull":           PURPLE_DULL,
		"Deep Dark Green":       GREEN_DEEPDARK,
		"White Pink":            WHITEPINK,
		"White Blue":            WHITEBLUE,
		"Green White":           GREEN_WHITE,
		"Grey Blue":             GREY_BLUE,
		"Sand":                  SAND,
		"Peach":                 PEACH,
	}

	sentences := map[string]string{
		"Normal":                "This is normal.",
		"Bold":                  "This is bold.",
		"Dim":                   "This is dim.",
		"Underline":             "This is underlined.",
		"No Underline":          "This has no underline.",
		"High Intensity Blue":   "Blue Blue Blue BLUE_HI.",
		"High Intensity Green":  "Green Green Green GREEN_HI.",
		"High Intensity Purple": "Purple Purple Purple PURPLE_HI",
		"Light Grey":            "Grey Grey Light Grey",
		"Red":                   "Red RED red red red.",
		"Green":                 "Green green normal green.",
		"Blue":                  "Blue,,,,,Blue blue BLUE.",
		"Yellow":                "yellowyellowyellowyellow.",
		"Light Yellow":          "Light Yellow LIGHT YELLOW.",
		"Purple Dull":           "Purple, dull .",
		"Deep Dark Green":       "Green, dark",
		"White Pink":            "White Pinkish Tone",
		"White Blue":            "White Blue Tone",
		"Green White":           "Green White Tone",
		"Grey Blue":             "Grey Blue Tone",
		"Sand":                  "Sand Tone",
		"Peach":                 "Peach Tone",
		"Cool Blue":             "Cool Blue COOL BLUE.",
		"Deep Blue":             "Deep Blue DEEP BLUE.",
	}

	for colorName, colorCode := range colors {
		fmt.Printf("%s%s%s\n", colorCode, sentences[colorName], NC)
		fmt.Printf("%s%s%s (bold)%s\n", BOLD, colorCode, sentences[colorName], NC)
		fmt.Printf("%s%s%s (dim)%s\n", DIM, colorCode, sentences[colorName], NC)

	}
}

func HelpTest() {
	fmt.Printf("%svirsh%s : wrapper around the %sC library%s %slibvirt%s\n", GREEN, NC, BOLD, NC, GREEN_HI, NC)
	fmt.Printf("%svirsh%s : wrapper around the C library %slibvirt%s\n", PURP_HI, NC, GREEN_HI, NC)
	fmt.Printf("%svirsh%s : wrapper around the C library %slibvirt%s\n", BLUE_HI, NC, BLUE, NC)
	fmt.Printf("%s%svirsh%s : wrapper around the C library %slibvirt%s\n", BOLD, BLUE_HI, NC, BLUE, NC)

	fmt.Println(DIM + "This is dim text" + NC)
	fmt.Println(LIGHT_GREY + "This is light grey text" + NC)

	fmt.Println(LIGHTPINK + "This is light pink text" + NC)
	fmt.Println(WHITEPINK + "This is whiTe pink text" + NC)
	fmt.Println(COOLBLUE + "This is cool blue text" + NC)
	fmt.Println(GREENISH + "This is greenish text" + NC)
	fmt.Println(PURPCOOL + "This is purple cool text" + NC)

	fmt.Println(BLUEHYPER + "This is blue hyper text" + NC)
	fmt.Println(PURPLE + "This is purple text" + NC)
	fmt.Println(GREEN_WHITE + "This is green white text" + NC)
	fmt.Println(BOLD + GREEN_WHITE + "This is bold green white text" + NC)
	fmt.Println(GREEN_LIGHT + "This is green light text" + NC)
	fmt.Println(SAND + "This is sand text" + NC)
	fmt.Println(PURPLE_DULL + "This is purple dull text" + NC)
	fmt.Println(GREEN_DEEPDARK + "This is green deep dark text" + NC)
	fmt.Println(TICK_GREEN + " " + BOLD + "Commands Successful" + NC)
	fmt.Println(CROSS_RED + " " + BOLD + "Failures Detected" + NC)

	fmt.Printf("%s\n%s%s%s\n%s\n", SECTION, GREY_BLUE, "Section Steps Running", NC, SECTION)
	fmt.Printf("%s\n%s%s%s%s\n%s\n", SECTION, BOLD, GREY_BLUE, "SECTION STEPS RUNNING", NC, SECTION)

	LogMainAction("1. STEP A")
	LogSection("RUNNING KUBE MASTER")
	LogStep("Step 1. Performing xyz")

	LogError("Failed to Launch Cluster")

	fmt.Printf("%s%s%s", LIGHT_YELLOW, "LIGHT Warning Message: Be Careful", NC)
	fmt.Printf("%s%s%s%s", DIM, LIGHT_YELLOW, "LIGHT Warning Message: Be Careful", NC)
	fmt.Printf("%s%s%s", YELLOW, "Warning Message: Be Careful", NC)
	fmt.Printf("%s%s%s%s", DIM, YELLOW, "Warning Message: Be Careful", NC)
}

// VM=mydomain sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml $VM | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'
