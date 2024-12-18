package main

import (
	"fmt"
	"hangman/game" // Assurez-vous que le nom du module est correct
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type GameSession struct {
	Pseudo       string
	Difficulty   string
	Word         string
	Attempts     int
	HiddenWord   string
	TriedLetters []string
	Message      string
	HangmanImage string
}

var gameSession *GameSession
var userSession map[string]string = make(map[string]string)

const scoreFilePath = "scores.txt"

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/start", startGameHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/guess", guessHandler)
	http.HandleFunc("/end", endGameHandler)
	http.HandleFunc("/scores", scoresHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/regles", handleRegles)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pseudo := r.FormValue("name") // Vérifiez que "name" correspond bien au nom du champ input
		if pseudo == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Stocker le pseudo dans la session
		gameSession = &GameSession{Pseudo: pseudo}

		// Redirection vers la page des règles
		http.Redirect(w, r, "/regles", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de chargement de la page d'accueil", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pseudo := r.FormValue("pseudo")
		difficulty := r.FormValue("difficulty")

		word, hiddenWord, attempts := game.NewGame(difficulty)

		gameSession = &GameSession{
			Pseudo:       pseudo,
			Difficulty:   difficulty,
			Word:         word,
			Attempts:     attempts,
			HiddenWord:   hiddenWord,
			TriedLetters: []string{},
			Message:      "",
			HangmanImage: "/static/images/hangman0.png",
		}

		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	if gameSession == nil || gameSession.Attempts <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tmpl, err := template.ParseFiles("templates/game.html")
	if err != nil {
		http.Error(w, "Erreur de chargement de la page de jeu", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, gameSession)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && gameSession != nil {
		guess := r.FormValue("guess")

		if !isValidInput(guess) {
			gameSession.Message = "Veuillez entrer uniquement des lettres."
		} else {
			updatedHiddenWord, attemptsLeft, message, triedLetters, hangmanImage := game.MakeGuess(
				gameSession.Word, gameSession.HiddenWord, guess, gameSession.Attempts, gameSession.TriedLetters)

			gameSession.HiddenWord = updatedHiddenWord
			gameSession.Attempts = attemptsLeft
			gameSession.Message = message
			gameSession.TriedLetters = triedLetters
			gameSession.HangmanImage = hangmanImage

			if attemptsLeft == 0 || updatedHiddenWord == gameSession.Word {
				saveScore(gameSession.Pseudo, gameSession.Attempts > 0)
				http.Redirect(w, r, "/end", http.StatusSeeOther)
				return
			}
		}
	}
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func isValidInput(input string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z]+$", input)
	return match
}
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Récupérer le pseudo du formulaire
		pseudo := r.FormValue("name")

		// Stocker le pseudo dans une "session" ou une structure de données
		userSession["pseudo"] = pseudo

		// Rediriger vers la page des règles du jeu
		http.Redirect(w, r, "/regles", http.StatusFound)
		return
	}

	// Sinon, on affiche la page d'accueil
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func handleRegles(w http.ResponseWriter, r *http.Request) {
	if gameSession == nil || gameSession.Pseudo == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/regle.html")
	if err != nil {
		http.Error(w, "Erreur de chargement de la page des règles", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, gameSession)
}

func endGameHandler(w http.ResponseWriter, r *http.Request) {
	if gameSession == nil || gameSession.Attempts > 0 {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	messages := []string{
		"Félicitations ! Vous avez gagné !",
		"Bravo, vous avez trouvé le mot !",
		"Malheureusement, c'est une défaite.",
		"Pas de chance cette fois ! Essayez encore.",
	}
	rand.Seed(time.Now().Unix())
	message := messages[rand.Intn(len(messages))]

	tmpl, err := template.ParseFiles("templates/end.html")
	if err != nil {
		http.Error(w, "Erreur de chargement de la page de fin", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, struct {
		Message string
		Success bool
	}{Message: message, Success: gameSession.Attempts > 0})

	gameSession = nil // Réinitialise la session de jeu
}

func saveScore(pseudo string, won bool) {
	result := "perdu"
	if won {
		result = "gagné"
	}
	entry := fmt.Sprintf("%s - %s\n", pseudo, result)
	f, err := os.OpenFile(scoreFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Erreur lors de la sauvegarde des scores:", err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	if err != nil {
		log.Println("Erreur lors de l'écriture dans le fichier:", err)
	}
}

func scoresHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(scoreFilePath)
	if err != nil {
		log.Println("Erreur de lecture des scores:", err)
		data = []byte("Aucun score disponible.")
	}
	scores := strings.Split(string(data), "\n")
	tmpl, err := template.ParseFiles("templates/scores.html")
	if err != nil {
		http.Error(w, "Erreur de chargement de la page des scores", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, scores)
}
