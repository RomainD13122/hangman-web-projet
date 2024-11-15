package main

import (
	"fmt"
	"hangman/game"
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

const scoreFilePath = "scores.txt"

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/start", startGameHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/guess", guessHandler)
	http.HandleFunc("/end", endGameHandler)
	http.HandleFunc("/scores", scoresHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if gameSession != nil && gameSession.Attempts > 0 {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
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
	tmpl := template.Must(template.ParseFiles("templates/game.html"))
	tmpl.Execute(w, gameSession)
}
func MakeGuess(word, hiddenWord, guess string, attempts int, triedLetters []string) (string, int, string, []string, string) {
	message := ""
	hangmanImage := "/static/images/hangman" + string(9-attempts) + ".png" // Image correspondante au nombre d'essais restants
	updatedHiddenWord := hiddenWord

	// Vérifier si le joueur a déjà tenté cette lettre
	if contains(triedLetters, guess) {
		message = "Vous avez déjà essayé cette lettre."
		return updatedHiddenWord, attempts, message, triedLetters, hangmanImage
	}

	// Ajouter la lettre à la liste des lettres essayées
	triedLetters = append(triedLetters, guess)

	// Si l'entrée est un mot entier
	if len(guess) > 1 {
		if guess == word {
			updatedHiddenWord = word
			message = "Félicitations, vous avez deviné le mot !"
			attempts = 0 // Le joueur a deviné le mot, donc plus de tentatives restantes
		} else {
			message = "Ce n'est pas le bon mot. Vous perdez 3 vies."
			attempts -= 3
		}
	} else {
		// Si l'entrée est une lettre
		letter := rune(guess[0])
		if containsRune(word, letter) {
			updatedHiddenWord = revealAllLetters(word, updatedHiddenWord, letter)
			message = "Bien joué !"
		} else {
			message = "Cette lettre n'est pas dans le mot."
			attempts--
		}
	}

	return updatedHiddenWord, attempts, message, triedLetters, hangmanImage
}

// Fonction utilitaire pour vérifier si une lettre ou un mot est dans la liste des lettres déjà essayées
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// Fonction utilitaire pour vérifier si une rune (lettre) est dans un mot
func containsRune(word string, letter rune) bool {
	for _, l := range word {
		if l == letter {
			return true
		}
	}
	return false
}

// Fonction pour révéler les lettres dans le mot
func revealAllLetters(word, hiddenWord string, letter rune) string {
	updatedWord := ""
	for i := 0; i < len(word); i++ {
		if rune(word[i]) == letter {
			updatedWord += string(word[i])
		} else {
			updatedWord += string(hiddenWord[i]) // Garder les caractères déjà révélés
		}
	}
	return updatedWord
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

// Gère la page de fin de partie
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

	tmpl := template.Must(template.ParseFiles("templates/end.html"))
	tmpl.Execute(w, struct {
		Message string
		Success bool
	}{Message: message, Success: gameSession.Attempts > 0})

	gameSession = nil // Réinitialise la session de jeu
}

// Sauvegarde le score dans un fichier texte
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

// Affiche la page de scores
func scoresHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(scoreFilePath)
	if err != nil {
		log.Println("Erreur de lecture des scores:", err)
		data = []byte("Aucun score disponible.")
	}
	scores := strings.Split(string(data), "\n")
	tmpl := template.Must(template.ParseFiles("templates/scores.html"))
	tmpl.Execute(w, scores)
}
