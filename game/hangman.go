package game

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	for {
		playGame()

		// Demander à l'utilisateur s'il souhaite rejouer
		fmt.Print("Voulez-vous rejouer ? (Appuyez sur '+' pour rejouer, * pour quitter) : ")
		var input string
		fmt.Scan(&input)

		// Si l'utilisateur ne choisit pas '+', quitter le jeu
		if input != "+" {
			fmt.Println("Merci d'avoir joué ! À bientôt.")
			break
		}

		clearScreen() // Effacer l'écran pour une nouvelle partie
	}
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := game.RenderTemplate(w, "index.html", nil)
	if err != nil {
		log.Println("Erreur dans homeHandler:", err)
	}
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Logique pour démarrer le jeu, puis rediriger vers /game
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	// Logique de rendu du jeu
	err := game.RenderTemplate(w, "game.html", nil)
	if err != nil {
		log.Println("Erreur dans gameHandler:", err)
	}
}

func endGameHandler(w http.ResponseWriter, r *http.Request) {
	// Logique de fin de jeu et affichage de la page de fin
	err := game.RenderTemplate(w, "end.html", nil)
	if err != nil {
		log.Println("Erreur dans endGameHandler:", err)
	}
}

func NewGame(difficulty string) (string, string, int) {
	rand.Seed(time.Now().UnixNano()) // Initialise la graine du générateur aléatoire avec l'heure actuelle
	var word string
	var hiddenWord string
	var attempts int

	// Sélection du mot en fonction de la difficulté
	switch difficulty {
	case "easy":
		word = "chat"      // Mot facile
		hiddenWord = "___" // Mot masqué avec des underscores
		attempts = 10      // Nombre d'essais pour un mot facile
	case "medium":
		word = "éléphant"       // Mot de difficulté moyenne
		hiddenWord = "________" // Mot masqué pour un mot de taille moyenne
		attempts = 7            // Nombre d'essais pour un mot de difficulté moyenne
	case "hard":
		word = "hippopotame"      // Mot difficile
		hiddenWord = "__________" // Mot masqué pour un mot plus long
		attempts = 5              // Nombre d'essais pour un mot difficile
	default:
		word = "exemple"       // Mot par défaut si aucune difficulté n'est spécifiée
		hiddenWord = "_______" // Mot masqué par défaut
		attempts = 9           // Nombre d'essais par défaut
	}

	// Retourner le mot, le mot masqué et le nombre d'essais
	return word, hiddenWord, attempts
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

// Fonction utilitaire pour vérifier si un élément existe dans une tranche de chaînes
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

// Fonction qui contient le jeu du pendu
func playGame() {
	// Demander à l'utilisateur de choisir une difficulté
	var choix int
	fmt.Println("Entrez une difficulté :")
	fmt.Println("(1) Facile")
	fmt.Println("(2) Difficile")
	fmt.Scan(&choix)

	var fileName string
	switch choix {
	case 1:
		fileName = "hangman.txt"
	case 2:
		fileName = "hangman1.txt"
	default:
		fmt.Println("Choix invalide. Utilisation du fichier par défaut : hangman.txt.")
		fileName = "hangman.txt"
	}

	// Lecture des mots depuis le fichier choisi
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Impossible d'ouvrir le fichier : %v", err)
	}
	defer file.Close()

	var mots []string
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		mots = append(mots, strings.Fields(line)...)
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Erreur lors de la lecture du fichier : %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	// Sélectionner un mot aléatoire dans la liste
	motAleatoire := mots[rand.Intn(len(mots))]

	// Affichage initial : révéler une ou deux lettres selon la longueur du mot
	var affichage string
	if len(motAleatoire) >= 10 {
		indicesVisibles := rand.Perm(len(motAleatoire))[:2] // Deux indices aléatoires
		affichage = replaceWithMultipleLetters(motAleatoire, indicesVisibles)
	} else {
		lettreVisible := rune(motAleatoire[rand.Intn(len(motAleatoire))])
		affichage = replaceWithUnderscores(motAleatoire, lettreVisible)
	}

	// Nombre de vies
	vie := 9

	// Boucle principale pour deviner le mot
	for vie > 0 {
		clearScreen() // Ajouter l'appel ici pour effacer l'écran avant chaque affichage

		// Afficher le dessin du pendu
		fmt.Println(drawHangman(vie))

		// Afficher le mot avec les lettres visibles
		fmt.Printf("Le mot à deviner est : %s\n", affichage)
		fmt.Printf("Il vous reste %d vies.\n", vie)

		// Demander à l'utilisateur d'entrer une lettre ou un mot entier
		fmt.Print("Entrez une lettre ou le mot complet (* pour quitter) : ")
		var input string
		fmt.Scan(&input)

		// Permettre de quitter le jeu avec *
		if input == "*" {
			fmt.Println("Vous avez quitté le jeu.")
			break
		}

		// Vérifier si l'utilisateur tente de deviner le mot entier
		if len(input) > 1 {
			if input == motAleatoire {
				fmt.Printf("Félicitations, vous avez deviné le mot : %s\n", motAleatoire)
				break
			} else {
				fmt.Println("Ce n'est pas le bon mot ! Vous perdez deux vies.")
				vie -= 2 // Perdre deux vies si le mot est incorrect
				continue
			}
		}

		// Si l'utilisateur a entré une seule lettre
		if len(input) == 1 {
			lettreDevinee := rune(input[0]) // Convertir la lettre en rune

			// Mettre à jour l'affichage si la lettre est correcte
			if containsRune(motAleatoire, lettreDevinee) {
				affichage = revealAllLetters(motAleatoire, affichage, lettreDevinee)
				fmt.Println("Bien joué !")
			} else {
				fmt.Println("Ce mot ne contient pas cette lettre.")
				vie-- // Perdre une vie si la lettre n'est pas dans le mot
			}
		} else {
			fmt.Println("Veuillez entrer une seule lettre ou un mot complet.")
		}

		// Vérifier si le mot est complètement deviné
		if !containsUnderscore(affichage) {
			fmt.Printf("Félicitations, vous avez deviné le mot : %s\n", motAleatoire)
			break
		}
		// Vérifier si les vies sont épuisées
		if vie <= 0 {
			fmt.Printf("Vous avez perdu. Le mot était : %s\n", motAleatoire)
			break
		}
	}

	clearScreen() // Effacement de l'écran après la fin de la partie
}
func containsInt(slice []int, item int) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
func replaceWithMultipleLetters(word string, indices []int) string {
	affichage := ""
	for i := 0; i < len(word); i++ {
		if containsInt(indices, i) {
			affichage += string(word[i]) // Garder la lettre si l'indice est dans le tableau
		} else {
			affichage += "_" // Remplacer par un underscore si l'indice n'est pas dans le tableau
		}
	}
	return affichage
}

// Fonction pour dessiner le pendu en fonction du nombre de vies restantes
func drawHangman(vie int) string {
	steps := []string{
		`
			
			  | 
			  | 
			  | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
			  |     
			  | 
			  | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
			  | 
			  | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
			  | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
		  |   | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
		 /|   | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
		 /|\  | 
			  | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
		 /|\  | 
		 /    | 
			  | 
		  ___/ `,
		`
		  ___  
		  |   | 
		  o   | 
		 /|\  | 
		 / \  | 
			  | 
		  ___/ `,
	}

	index := 9 - vie // Calculer l'étape du dessin à partir des vies restantes
	if index >= len(steps) {
		index = len(steps) - 1 // Assurer de ne pas dépasser les étapes
	}

	return steps[index]
}

// Fonction pour remplacer les lettres restantes par des underscores
func replaceWithUnderscores(mot string, lettreVisible rune) string {
	affichage := ""
	for _, lettre := range mot {
		if lettre == lettreVisible {
			affichage += string(lettre)
		} else {
			affichage += "_"
		}
	}
	return affichage
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func containsUnderscore(affichage string) bool {
	for _, l := range affichage {
		if l == '_' {
			return true
		}
	}
	return false
}
