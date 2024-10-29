package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
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

// Fonction qui contient le jeu du pendu
func playGame() {
	// Demander à l'utilisateur de choisir une difficulté
	var choix int
	fmt.Println("Entrez une difficulté :")
	fmt.Println("(1) Facile - hangman.txt")
	fmt.Println("(2) Difficile - hangman1.txt")
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

// Fonction pour afficher deux lettres aléatoires
func replaceWithMultipleLetters(mot string, indices []int) string {
	affichage := ""
	for i := 0; i < len(mot); i++ {
		if contains(indices, i) {
			affichage += string(mot[i])
		} else {
			affichage += "_"
		}
	}
	return affichage
}

// Fonction pour vérifier si un indice est dans une liste
func contains(indices []int, val int) bool {
	for _, index := range indices {
		if index == val {
			return true
		}
	}
	return false
}

// Fonction pour effacer l'écran selon l'OS
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// Fonction pour vérifier si une rune est dans un mot
func containsRune(mot string, lettre rune) bool {
	for _, l := range mot {
		if l == lettre {
			return true
		}
	}
	return false
}

// Fonction pour dévoiler toutes les lettres correspondantes dans l'affichage
func revealAllLetters(mot string, affichage string, lettre rune) string {
	newAffichage := ""
	for i := 0; i < len(mot); i++ {
		if rune(mot[i]) == lettre {
			newAffichage += string(mot[i])
		} else {
			newAffichage += string(affichage[i]) // garder le même caractère
		}
	}
	return newAffichage
}

// Fonction pour vérifier s'il reste des underscores
func containsUnderscore(affichage string) bool {
	for _, l := range affichage {
		if l == '_' {
			return true
		}
	}
	return false
}
