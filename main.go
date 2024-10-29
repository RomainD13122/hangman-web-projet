package main

package main

import (
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "regexp"
    "strings"
    "time"
    "hangman-web/game" // Importer le package contenant le jeu Hangman (votre logique CLI)
)

type GameSession struct {
    Pseudo        string
    Difficulty    string
    Word          string
    Attempts      int
    HiddenWord    string
    TriedLetters  []string
    Message       string
    HangmanImage  string
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