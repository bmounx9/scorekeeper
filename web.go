package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"github.com/gosimple/slug"
	"encoding/json"
	"strings"
)

type Page struct {
	Title        string
	DisplayTitle string
	Metadata     map[string]string
	HTMLMetadata map[string]template.HTML
}

// load game data
// TODO: currently loading from JSON files, switch to persistence layer
func loadGameData(title string) (*Page, error) {
	// read data if it exists
	filename := "data/game-" + slug.Make(title) + ".json"
	encoded, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// decode data into page struct
	p := &Page{}
	decodeError := json.Unmarshal(encoded, &p)

	return p, decodeError
}

// save game data
func (p *Page) saveGameData() error {
	filename := "data/game-" + slug.Make(p.Title) + ".json"
	encoded, _ := json.Marshal(p)
	return ioutil.WriteFile(filename, encoded, 0600)
}

// render form for creating game with data filled in if editing
func createGameHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadGameData(title)
	if err != nil {
		p = &Page{}
	}
	renderTemplate(w, "create-game", p)
}

// save game data and redirect to view if successful
func saveGameHandler(w http.ResponseWriter, r *http.Request, title string) {
	// handle title
	displayTitle := r.FormValue("title")
	slugTitle := slug.Make(displayTitle)

	// gather metadata
	meta := make(map[string]string)
	meta["description"] = r.FormValue("description")

	// create page and save
	p := &Page{Title: slugTitle, DisplayTitle: displayTitle, Metadata: meta}
	err := p.saveGameData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// redirect to view
	http.Redirect(w, r, "/view-game/"+slugTitle, http.StatusFound)
}

// view game data
func viewGameHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadGameData(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// format scores
	scoresString := "<table style=\"border: 1px solid black;\">"
	if p.Metadata["scores"] != "" {
		scores := strings.Split(p.Metadata["scores"], "\n")
		for _, s := range scores {
			if s != "" {
				scoreData := strings.Split(s, ",")
				scoresString += "<tr><td>" + scoreData[0] + "</td><td>" + scoreData[1] + "</td><td>" + scoreData[2] + "</td><td>" + scoreData[3] + "</td></tr>"
			}
		}
	}
	scoresString += "</table>"

	// add to view
	htmlMetadata := make(map[string]template.HTML)
	htmlMetadata["scores"] = template.HTML(scoresString)
	p.HTMLMetadata = htmlMetadata

	renderTemplate(w, "view-game", p)
}

// add score data
func addScoreHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadGameData(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// build user list
	htmlMetadata := make(map[string]template.HTML)
	htmlMetadata["users"] = template.HTML(getUserOption())

	p.HTMLMetadata = htmlMetadata
	renderTemplate(w, "add-score", p)
}

// get list of users as an HTML select
func getUserOption() (string) {
	files, err := ioutil.ReadDir("data/")
	if err != nil {
		log.Fatal(err)
	}

	users := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "user-") {
			name := strings.Replace(strings.Replace(f.Name(), ".json", "", -1), "user-", "", -1)
			p, _ := loadUserData(name)
			users += `<option value="` + name + `">` + p.DisplayTitle + `</option>`
		}
	}

	return users
}

// save game data and redirect to view if successful
func saveScoreHandler(w http.ResponseWriter, r *http.Request, title string) {
	// load existing data
	p, err := loadGameData(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// add score and save data
	p.Metadata["scores"] += r.FormValue("player-one") + "," + r.FormValue("score-one") + "," + r.FormValue("player-two") + "," + r.FormValue("score-two") + "\n"
	saveErr := p.saveGameData()
	if saveErr != nil {
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}

	// redirect to view
	http.Redirect(w, r, "/view-game/"+p.Title, http.StatusFound)
}

// load user data
// TODO: currently loading from JSON files, switch to persistence layer
func loadUserData(title string) (*Page, error) {
	// read data if it exists
	filename := "data/user-" + slug.Make(title) + ".json"
	encoded, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// decode data into page struct
	p := &Page{}
	decodeError := json.Unmarshal(encoded, &p)

	return p, decodeError
}

// save game data
func (p *Page) saveUserData() error {
	filename := "data/user-" + slug.Make(p.Title) + ".json"
	encoded, _ := json.Marshal(p)
	return ioutil.WriteFile(filename, encoded, 0600)
}

// render form for creating game with data filled in if editing
func createUserHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadUserData(title)
	if err != nil {
		p = &Page{}
	}
	renderTemplate(w, "create-user", p)
}

// save game data and redirect to view if successful
func saveUserHandler(w http.ResponseWriter, r *http.Request, title string) {
	// handle title
	displayTitle := r.FormValue("title")
	slugTitle := slug.Make(displayTitle)

	// gather metadata
	meta := make(map[string]string)
	meta["description"] = r.FormValue("description")

	// create page and save
	p := &Page{Title: slugTitle, DisplayTitle: displayTitle, Metadata: meta}

	err := p.saveUserData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view-user/"+slugTitle, http.StatusFound)
}

// view game data
func viewUserHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadUserData(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "view-user", p)
}

// load report data
func loadReport() (*Page, error) {
	// build game list
	htmlMetadata := make(map[string]template.HTML)
	htmlMetadata["games"] = template.HTML(getGameList())
	htmlMetadata["users"] = template.HTML(getUserList())

	return &Page{Title: "report", HTMLMetadata: htmlMetadata}, nil
}

// get list of games as an HTML list
func getGameList() (string) {
	files, err := ioutil.ReadDir("data/")
	if err != nil {
		log.Fatal(err)
	}

	games := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "game-") {
			name := strings.Replace(strings.Replace(f.Name(), ".json", "", -1), "game-", "", -1)
			p, _ := loadGameData(name)
			games += `<li><a href="/view-game/` + name + `">` + p.DisplayTitle + `</a></li>`
		}
	}

	return games
}

// get list of users as an HTML list
func getUserList() (string) {
	files, err := ioutil.ReadDir("data/")
	if err != nil {
		log.Fatal(err)
	}

	users := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "user-") {
			name := strings.Replace(strings.Replace(f.Name(), ".json", "", -1), "user-", "", -1)
			p, _ := loadUserData(name)
			users += `<li><a href="/view-user/` + name + `">` + p.DisplayTitle + `</a></li>`
		}
	}

	return users
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadReport()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "report", p)
}

// validation for templates
var templates = template.Must(
	template.ParseFiles(
		"template/report.html",
		"template/create-game.html",
		"template/view-game.html",
		"template/add-score.html",
		"template/create-user.html",
		"template/view-user.html",
	))

func renderTemplate(w http.ResponseWriter, template string, p *Page) {
	err := templates.ExecuteTemplate(w, template+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// validation for paths
var validPath = regexp.MustCompile("^/(create-game|edit-game|view-game|save-game|add-score|save-score|create-user|edit-user|view-user|save-user)/?([a-zA-Z0-9-]+)?$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	// home page
	http.HandleFunc("/", reportHandler)

	// game routes
	http.HandleFunc("/create-game/", makeHandler(createGameHandler))
	http.HandleFunc("/edit-game/", makeHandler(createGameHandler))
	http.HandleFunc("/view-game/", makeHandler(viewGameHandler))
	http.HandleFunc("/save-game/", makeHandler(saveGameHandler))
	http.HandleFunc("/add-score/", makeHandler(addScoreHandler))
	http.HandleFunc("/save-score/", makeHandler(saveScoreHandler))

	// user routes
	http.HandleFunc("/create-user/", makeHandler(createUserHandler))
	http.HandleFunc("/edit-user/", makeHandler(createUserHandler))
	http.HandleFunc("/view-user/", makeHandler(viewUserHandler))
	http.HandleFunc("/save-user/", makeHandler(saveUserHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
