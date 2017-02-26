/* 
Saezuri, a Go + ebiten toy implementation of a GNU Social client
MIT License Copyright (c) 2017 Myu-Unix 
*/

package main

import (
        _ "image/jpeg"
        _ "image/png"
        "log"
        "fmt"
	"bufio"
	"os"
	"strings"
	"os/exec"
	"time"
	"encoding/xml"
        "github.com/hajimehoshi/ebiten"
        "github.com/hajimehoshi/ebiten/ebitenutil"
)

type GNUS_XML struct {
	Statuses	string		`xml:"statuses"`
	StatusList	[]Status	`xml:"status"`
}
type Status struct {
	Text	string	`xml:"text"`
	Id	string	`xml:"id"`
	Time	string	`xml:"created_at"`
	Source	string	`xml:"source"`
	FavNum	string  `xml:"fave_num"`
	UserList	[]User	`xml:"user"`
}
type User struct {
	UserName	string	`xml:"name"`
	ProfileImg	string	`xml:"profile_image_url"`
}

var keyNames = map[ebiten.Key]string{
        ebiten.KeyTab:       "@", // ebiten doesn't seems to support special keys
}

var (
	user string
	pwd string
	config_file = "saezuri.conf"
	instance_url string
	SurfaceWallpaper *ebiten.Image
	SurfaceDarkOverlay *ebiten.Image
	SurfaceLogo *ebiten.Image
	SurfaceKeybindings *ebiten.Image
	SurfaceNewNotice *ebiten.Image
	notice [20]string // 20 notices max are returned
	pressed []string
	strArray []string
	location string
	timestamp string
	msgstr string
	menu_saezuri_message = "A Go + ebiten GNU Social client\n"
	menu_saezuri_url = "https://github.com/saezuri-eb\n"
	menu_keybindings_title = "Keybindings :\n"
	menu_keybind1 = "   Shift+H: Show home timeline\n"
	menu_keybind2 = "   Shift+O: Show own timeline\n"
	menu_keybind3 = "   Shift+N: New Notice\n"
	menu_keybind4 = "   Shift+A: Show mentions\n"
	menu_keybind5 = "   Shift+M: Go back to this menu\n"
	menu_keybind6 = "   Shift+Q: Quit\n"
	space_allowed = 1 // uglyyy
	show = 1 // Initial value, will display splash
	called = 0
	goroutine_launched = 0
	i = 0
	id = 0 // For notice details
	cmdOut []byte
	err    error
	notice_id int
	keyStates = map[ebiten.Key]int{}
	message string
	api_action string
	// curl request basic variables
	app = "curl"
	arg1 string
	args []string
)

func read_config() error {
fmt.Printf("Searching for config file under %s\n", config_file)
// read user,pwd & instance_url from config_file
file, err := os.Open(config_file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    i = 0

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
	if i == 0 {
        user = scanner.Text()
	}
	if i == 1 {
        pwd = scanner.Text()
	}
	if i == 2 {
        instance_url = scanner.Text()
	}
	i++
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
fmt.Printf("Username: %s\n", user)
fmt.Printf("Instance URL: %s\n", instance_url)

// generate the user:pwd line for cURL
arg1 = fmt.Sprintf("%s:%s", user, pwd)

return nil
}

func display_splash2(screen *ebiten.Image) error {
	// Display "slash screen" / main menu
        str := `{{.Padding}}  {{.SaezuriMessage}}  {{.SaezuriURL}}  {{.UserName}}@{{.Instance}} {{.Padding}}{{.Padding}}{{.Padding}} {{.KeyTitle}} {{.Key1}} {{.Key2}} {{.Key3}} {{.Key4}} {{.Key5}} {{.Key6}}`
	str = strings.Replace(str, "{{.Padding}}", "\n\n\n\n", -1)
	str = strings.Replace(str, "{{.SaezuriMessage}}", menu_saezuri_message, -1)
	str = strings.Replace(str, "{{.SaezuriURL}}", menu_saezuri_url, -1)
	str = strings.Replace(str, "{{.UserName}}", user, -1)
	str = strings.Replace(str, "{{.Instance}}", instance_url, -1)
	str = strings.Replace(str, "{{.KeyTitle}}", menu_keybindings_title, -1)
	str = strings.Replace(str, "{{.Key1}}", menu_keybind1, -1)
	str = strings.Replace(str, "{{.Key2}}", menu_keybind2, -1)
	str = strings.Replace(str, "{{.Key3}}", menu_keybind3, -1)
	str = strings.Replace(str, "{{.Key4}}", menu_keybind4, -1)
	str = strings.Replace(str, "{{.Key5}}", menu_keybind5, -1)
	str = strings.Replace(str, "{{.Key6}}", menu_keybind6, -1)
        if err := ebitenutil.DebugPrint(screen, str); err != nil {
                return err
        }
	return nil
}

func display_notices (screen *ebiten.Image) error {
	// Display notices, mentions, anything like a timeline :)
	if called == 0 {
	fmt.Printf("Calling API...\n")
        create_generic_call()
	fmt.Printf("Done\n")
	called = 1
	}

        str := `{{.Location}}{{.Timestamp}} ---{{.Notice0}} ---{{.Notice1}} ---{{.Notice2}} ---{{.Notice3}} ---{{.Notice4}} ---{{.Notice5}} ---`
	str = strings.Replace(str, "{{.Location}}", location, -1)
        str = strings.Replace(str, "{{.Notice0}}", notice[0], -1)
	str = strings.Replace(str, "{{.Notice1}}", notice[1], -1)
	str = strings.Replace(str, "{{.Notice2}}", notice[2], -1)
	str = strings.Replace(str, "{{.Notice3}}", notice[3], -1)
	str = strings.Replace(str, "{{.Notice4}}", notice[4], -1)
	str = strings.Replace(str, "{{.Notice5}}", notice[5], -1)
	str = strings.Replace(str, "{{.Timestamp}}", timestamp, -1)
        if err := ebitenutil.DebugPrint(screen, str); err != nil {
                return err
        }
	return nil
}

func write_notice (screen *ebiten.Image) error {
        str := `{{.Location}} {{.Info1}} {{.Info2}}`
	str = strings.Replace(str, "{{.Location}}", location, -1)
	str = strings.Replace(str, "{{.Info1}}", "\n\n  @: Tab", -1)
	str = strings.Replace(str, "{{.Info2}}", "\n  Esc: Quit writing", -1)
        if err := ebitenutil.DebugPrint(screen, str); err != nil {
                return err
        }
	
	message = strings.Join(pressed,"") // convert string slice to string
	//fmt.Printf("%v\n", message)

	// Handle keypresses
	for c := 'a'; c <= 'z'; c++ {
	 if ebiten.IsKeyPressed(ebiten.Key(c) - 'a' + ebiten.KeyA) {
	    keyStates[ebiten.Key(c)]++
	  } else {
	    keyStates[ebiten.Key(c)] = 0 
        }
	if IsKeyTriggered(ebiten.Key(c)) == true {
	pressed = append(pressed, string(c))
	space_allowed = 1 // Allow space only after a letter, not after a space
	}
	}

	/* Don't allow 2 consecutive space.
	 It is a WORKAROUND for multiple spaces being inputed even using IsKeyTriggered(ebiten.KeySpace)*/
	if space_allowed == 1 {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
	  keyStates[ebiten.KeySpace]++
	} else {
	    keyStates[ebiten.KeySpace] = 0 
        }
	if IsKeyTriggered(ebiten.KeySpace) == true {
	pressed = append(pressed, " ")
	space_allowed = 0
	}
	}
	
	// Handle "special keys"
   	for key, name := range keyNames {
        if ebiten.IsKeyPressed(key) {
	    keyStates[ebiten.KeyTab]++
	} else {
    	    keyStates[ebiten.KeyTab] = 0 
        }
	if IsKeyTriggered(ebiten.KeyTab) == true {
            pressed = append(pressed, name)
          }
        }

	// Delete last character
	if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
	  keyStates[ebiten.KeyBackspace]++
	} else {
	    keyStates[ebiten.KeyBackspace] = 0 
        }
	if IsKeyTriggered(ebiten.KeyBackspace) == true {
	size_pressed := len(pressed)
	if size_pressed > 0 {
	pressed = pressed[:size_pressed-1]
	}
	}

	// Post notice
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
  	keyStates[ebiten.KeyEnter]++
	} else {
	    keyStates[ebiten.KeyEnter] = 0 
        }
	if IsKeyTriggered(ebiten.KeyEnter) == true {
	fmt.Printf("Calling API...\n")
        create_post_call(message)
	// Clear dirty strings
	message = message[:0]
	msgstr = msgstr[:0]
	pressed = pressed[:0]
	// Go back to the home timeline after notice is posted
	api_action = "api/statuses/home_timeline.xml"
	location = "- Home Timeline "
	show = 2
	}

	// Handle keyboard Esc
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
	fmt.Printf("Esc ...\n")
	// Clear dirty strings
	message = message[:0]
	msgstr = msgstr[:0]
	pressed = pressed[:0]
	api_action = "api/statuses/home_timeline.xml"
	location = "- Home Timeline "
	show = 2
	}

	if len(message) < 48 {
        msgstr = ` {{.Padding}}{{.Message}}`
	msgstr = strings.Replace(msgstr, "{{.Padding}}", "\n\n\n\n\n\n\n\n\n\n\n\n     ", -1)
        msgstr = strings.Replace(msgstr, "{{.Message}}", message, -1)
	} else if len(message) > 48 && len(message) < 96 {
	msgstr = ` {{.Padding}}{{.Message}}{{.CR}}{{.Message2}}`
	msgfirstHalfSize := len(message)/2
    	msgfirstHalf := message[0:msgfirstHalfSize]
    	msgsecondHalf := message[msgfirstHalfSize:len(message)]
	msgstr = strings.Replace(msgstr, "{{.Padding}}", "\n\n\n\n\n\n\n\n\n\n\n\n     ", -1)
        msgstr = strings.Replace(msgstr, "{{.Message}}", msgfirstHalf, -1)
	msgstr = strings.Replace(msgstr, "{{.CR}}", "\n     ", -1)
	msgstr = strings.Replace(msgstr, "{{.Message2}}", msgsecondHalf, -1)
	}

	ebitenutil.DebugPrint(screen, msgstr)

	return nil
}

func IsKeyTriggered(key ebiten.Key) bool {
  return keyStates[key] == 1
}

func update_notice() {
	// This is a goroutine which starts & stops when displaying/not displaying notices
	for show == 2 {
	timestamp = fmt.Sprintf("Refreshed at %v\n\n", time.Now().Format(time.RFC822))
	fmt.Printf("update_notice goroutine: sleeping 120s\n")
	// Set a timer to enable refresh every X seconds
	time.Sleep(120* time.Second)
	called = 0
	}
	fmt.Printf("update_notice goroutine: Bye!\n")
	goroutine_launched = 0
}

func update(screen *ebiten.Image) error {
	i = 0
	screen.Clear()

	SurfaceWallpaperOp := &ebiten.DrawImageOptions{}
	SurfaceDarkOverlayOp := &ebiten.DrawImageOptions{}
	SurfaceLogoOp := &ebiten.DrawImageOptions{}
	SurfaceNewNoticeOp := &ebiten.DrawImageOptions{}
	
	// After screen is cleared, display Wallpaper
        if err := screen.DrawImage(SurfaceWallpaper, SurfaceWallpaperOp); err != nil {
              return err
        }

	 // If modifier key is pressed
	if ebiten.IsKeyPressed(ebiten.KeyShift) {

	// Handle keyboard H -> Home/friends
	if ebiten.IsKeyPressed(ebiten.KeyH) { 
	api_action = "api/statuses/home_timeline.xml"
	location = "- Home Timeline "
	show = 2
	called = 0
	}

	// Handle keyboard O -> Own
	if ebiten.IsKeyPressed(ebiten.KeyO) { 
	api_action = "api/statuses/user_timeline.xml"
	location = "- Own Timeline "
	show = 2
	called = 0
	}

	// Handle keyboard M -> Menu (splash)
	if ebiten.IsKeyPressed(ebiten.KeyM) { 
	show = 1
	called = 0
	}

	// Handle keyboard A -> Mentions
	if ebiten.IsKeyPressed(ebiten.KeyA) { 
	location = "- Mentions Timeline "
	api_action = "api/statuses/mentions.xml"
	show = 2
	called = 0
	}

	// Handle keyboard Q -> Exit
	if ebiten.IsKeyPressed(ebiten.KeyQ) { 
	os.Exit(0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyN) {
	// Handle keyboard N -> New
	location = "- Write new notice "
	api_action = "api/statuses/update.xml"
	show = 3
	called = 0
	}
	}

	// depending on the value of show, one of these functions is executed
	switch show {
	case 1:	// Display logo only on "splash"
		SurfaceLogoOp.GeoM.Translate(0, 15)
	      	if err := screen.DrawImage(SurfaceLogo, SurfaceLogoOp); err != nil {
        	      return err
		}
		display_splash2(screen)
	case 2: // Put dark background "under" notices
		timestamp = fmt.Sprintf("Refreshed at %v\n\n", time.Now().Format(time.RFC822))
      		if err := screen.DrawImage(SurfaceDarkOverlay, SurfaceDarkOverlayOp); err != nil {
              		return err
       		}
		if goroutine_launched == 0 {
		go update_notice() // Launch update goroutine
		goroutine_launched = 1
		}
		display_notices(screen)
	case 3:
		SurfaceNewNoticeOp.GeoM.Translate(10, 175)
		// Put dark background "under" notices
      		if err := screen.DrawImage(SurfaceDarkOverlay, SurfaceDarkOverlayOp); err != nil {
              		return err
       		}
		// Put overlay on top to write notice
      		if err := screen.DrawImage(SurfaceNewNotice, SurfaceNewNoticeOp); err != nil {
              		return err
       		}
		write_notice(screen)
	}
        return nil
}

func main() {
        var err error

	read_config()

	SurfaceWallpaper, _, err = ebitenutil.NewImageFromFile("images/background.jpg", ebiten.FilterNearest)
  	if err != nil {
                log.Fatal(err)
        }
	SurfaceDarkOverlay, _, err = ebitenutil.NewImageFromFile("images/dark_overlay.png", ebiten.FilterNearest)
  	if err != nil {
                log.Fatal(err)
        }
	SurfaceLogo, _, err = ebitenutil.NewImageFromFile("images/logo_on_dark.png", ebiten.FilterNearest)
  	if err != nil {
                log.Fatal(err)
        }
	SurfaceNewNotice, _, err = ebitenutil.NewImageFromFile("images/overlay_new_notice.png", ebiten.FilterNearest)
  	if err != nil {
                log.Fatal(err)
        }
        if err := ebiten.Run(update, 360, 480, 1.0, "Saezuri"); err != nil {
                log.Fatal(err)
        }
}

// *** API & XML part ***

func create_post_call(message string) {
    // create a API call to post a notice
    arg2 := fmt.Sprintf("%s/%s", instance_url, api_action)
    arg4 := fmt.Sprintf("status=%s", message)
    args := []string{"-u", arg1, arg2, "-d", arg4}
    fmt.Printf("api_call :%s %s\n", app, args) 
    api_generic_call(args)
}

func create_delete_call() {
    // create a API call to delete a notice
    arg2 := fmt.Sprintf("%s/%s", instance_url, api_action)
    args := []string{"-u", arg1, arg2, "-X", "POST"}
    api_generic_call(args)
}

func create_generic_call() {
    // create a "generic" API call (no special arguments)
    arg2 := fmt.Sprintf("%s/%s", instance_url, api_action)
    args := []string{"-u", arg1, arg2}
    api_generic_call(args)
}

func api_generic_call(args []string) {
    // execute the call with curl, grab output
    fmt.Printf("api_call :%s %s\n", app, args) // debug WARN : This display user/pwd on console
    if cmdOut, err := exec.Command(app, args...).Output(); err == nil {
      xml_parse(cmdOut)
    } else {
      fmt.Fprintln(os.Stderr, "error: ", err)
      os.Exit(1)
    }
}

func xml_parse(cmdOut []byte) {
    // reset "array" contents, is it the good way to do it ? It works
    i = 0
    for i < 20 {  
    notice[i] = ""
    i++
    }
    // parse XML output
    a := GNUS_XML{}
    err := xml.Unmarshal(cmdOut, &a)
    if err != nil { panic(err) }
    i = 0
    for _, entry := range a.StatusList {
    //fmt.Printf("- %s : %v\nfrom %s id %s at %s\navatar image : %s\n\n", entry.UserList[0].UserName, entry.Text, entry.Source, entry.Id, entry.Time, entry.UserList[0].ProfileImg)
    if len(entry.Text) >= 60 && len(entry.Text) < 120 {
    firstHalfSize := len(entry.Text)/2
    firstHalf := entry.Text[0:firstHalfSize]
    secondHalf := entry.Text[firstHalfSize:len(entry.Text)]
    notice[i] = fmt.Sprintf("\n %v\n %v\n by %s, favorited %s times\n", firstHalf, secondHalf, entry.UserList[0].UserName, entry.FavNum)

    } else if len(entry.Text) >= 120 { // Ok for 140 chars sized ones, no promises if longer
    firstThirdSize := len(entry.Text)/3
    firstThird := entry.Text[0:firstThirdSize]
    secondThird := entry.Text[firstThirdSize:firstThirdSize*2]
    lastThird := entry.Text[firstThirdSize*2:len(entry.Text)]

    notice[i] = fmt.Sprintf("\n %v\n %v\n %v\n by %s, favorited %s times\n", firstThird, secondThird, lastThird, entry.UserList[0].UserName, entry.FavNum)

    } else { // Small one-line entries
    notice[i] = fmt.Sprintf("\n %v\n by %s, favorited %s times\n", entry.Text, entry.UserList[0].UserName, entry.FavNum)
    }
    i++
    }
}