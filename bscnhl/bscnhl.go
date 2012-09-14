package bscnhl

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type Game struct {
	HomePlayer, AwayPlayer string
	HomeTeam, AwayTeam     string
	HomeScore, AwayScore   string
	HomeShots, AwayShots   string
	HomeHits, AwayHits     string
	Date                   time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/newgame", newgame)
	http.HandleFunc("/addgame", addgame)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Game").Order("-Date").Limit(10)
	games := make([]Game, 0, 10)
	if _, err := q.GetAll(c, &games); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := recentGamesTemplate.Execute(w, games); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var recentGamesTemplate = template.Must(template.ParseFiles("bscnhl/main.html"))

func newgame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	fmt.Fprint(w, newGameTemplateHTML)
}

const newGameTemplateHTML = `
<html>
	<body>
		<form action="/addgame" method="post">
			<div><input type="radio" name="side" value="home"></input> Home</div>
			<div><input type="radio" name="side" value="away"></input> Away</div>
			<div>Opponent: <input type="text" name="opponent"></input></div>
			<div>Home Team: <input type="text" name="hometeam"></input></div>
			<div>Away Team: <input type="text" name="awayteam"></input></div>
			<div>Home Goals: <input type="text" name="homegoals"></input></div>
			<div>Away Goals: <input type="text" name="awaygoals"></input></div>
			<div>Home Shots: <input type="text" name="homeshots"></input></div>
			<div>Away Shots: <input type="text" name="awayshots"></input></div>
			<div>Home Hits: <input type="text" name="homehits"></input></div>
			<div>Away Hits: <input type="text" name="awayhits"></input></div>
			<div><input type="submit" value="Add Game"></div>
		</form>
	</body>
</html>
`

func addgame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Game{
		HomeTeam:  r.FormValue("hometeam"),
		HomeScore: r.FormValue("homegoals"),
		HomeShots: r.FormValue("homeshots"),
		HomeHits:  r.FormValue("homehits"),
		AwayTeam:  r.FormValue("awayteam"),
		AwayScore: r.FormValue("awaygoals"),
		AwayShots: r.FormValue("awayshots"),
		AwayHits:  r.FormValue("awayhits"),
		Date:      time.Now(),
	}
	if r.FormValue("side") == "home" {
		g.AwayPlayer = r.FormValue("opponent")
		if u := user.Current(c); u != nil {
			g.HomePlayer = u.String()
		}
	} else {
		g.HomePlayer = r.FormValue("opponent")
		if u := user.Current(c); u != nil {
			g.AwayPlayer = u.String()
		}
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Game", nil), &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
