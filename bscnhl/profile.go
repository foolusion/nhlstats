package bscnhl

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"net/http"
)

type Profile struct {
	Account      string
	Name         string
	FavoriteTeam string
	Friends      []string
}

// TODO: change profiles to update the user data everytime by searching the
// games they played. Calculate their stats from the profile data and show 
// the five most recent games.
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
	k := datastore.NewKey(c, "Profile", u.Email, 0, nil)
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
	if _, err := datastore.Put(c, k, &prof); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, prof, games)
}
