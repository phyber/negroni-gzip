# gzip

Gzip middleware for Negroni.
Mostly a copy of the Martini gzip module with small changes to make it function under Negroni.

Not tested very much but Works For Me (TM).

## Usage

~~~ go
import (
  "github.com/codegangsta/negroni"
  "github.com/phyber/negroni-gzip"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  }
  n := negroni.Classic()
  n.Use(gzip.Gzip())
  n.UseHandler(mux)
  n.Run(":3000")
}

~~~

Make sure to include the Gzip middleware above other middleware that alter the response body.

## Authors
* [Jeremy Saenz](http://github.com/codegangsta)
* [Shane Logsdon](http://github.com/slogsdon)
* [David O'Rourke](https://github.com/phyber)
