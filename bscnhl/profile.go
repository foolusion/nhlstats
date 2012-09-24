package bscnhl

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"html/template"
	"net/http"
	"sort"
)

type Games []Game

func (s Games) Len() int      { return len(s) }
func (s Games) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByDate struct{ Games }
type ByDateDesc struct{ Games }

func (s ByDate) Less(i, j int) bool { return s.Games[i].Date.Before(s.Games[j].Date) }

func (s ByDateDesc) Less(i, j int) bool { return s.Games[i].Date.After(s.Games[j].Date) }

type Profile struct {
	Account      string
	Name         string
	FavoriteTeam string
	Friends      []string
}

type profileindex struct {
	User         string
	Login        bool
	URL          string
	FavoriteTeam string
	Friends      []string
	Games        []Game
}

func profile(w http.ResponseWriter, r *http.Request) {
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
	var prof Profile
	games := make([]Game, 0, 20)
	prof.Account = u.Email
	q := datastore.NewQuery("Game").Filter("HomePlayer =", u.Email)
	for t := q.Run(c); ; {
		var g Game
		_, err := t.Next(&g)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		games = append(games, g)
	}
	q = datastore.NewQuery("Game").Filter("AwayPlayer =", u.Email)
	for t := q.Run(c); ; {
		var g Game
		_, err := t.Next(&g)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		games = append(games, g)
	}

	sort.Sort(ByDateDesc{games})

	logoutURL, err := user.LogoutURL(c, "/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	i := profileindex{
		User:         u.Email,
		Login:        false,
		URL:          logoutURL,
		FavoriteTeam: prof.FavoriteTeam,
		Friends:      prof.Friends,
		Games:        games,
	}

	if err := profileTemplate.Execute(w, i); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var profileTemplate = template.Must(template.ParseFiles("bscnhl/main.html", "bscnhl/gamelist.html"))
