package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	private  = make(chan string)

	// Variável que armazena os ponteiros com base na chave. Funciona como se fosse um vetor de ponteiros
	// e os indices são os nomes. Ex: uprivate[Ribas] retorna o ponteiro do nome Ribas
	uprivate = make(map[string]client)
)

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados

	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			for cli := range clients {
				cli <- msg
			}

		case msg := <-private:
			// broadcast de mensagens privadas. Envia para a pessoa escolhida
			// Split utilizado para pegar o nick do usuario para usar como chave no vetor uprivate
			nick := strings.Split(msg, " ")
			// Split utilizado para pegar o texto enviado
			texto := strings.SplitAfterN(msg, " ", 4)
			uprivate[nick[2]] <- nick[0] + ": " + texto[3]

		case cli := <-entering:
			// Entrada do cliente/usuário
			clients[cli] = true

		case cli := <-leaving:
			// Saida do cliente/usuário
			delete(clients, cli)
			close(cli)
		}
	}
}

// Função que inverte a string fornecida
func reverse_string(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	apelido := conn.RemoteAddr().String()

	ch <- "vc é " + apelido
	messages <- apelido + " chegou!"
	entering <- ch

	// Armazena no vetor uprivate na possição "Nome do Usuário" o ponteiro do usuário
	uprivate[apelido] = ch

	input := bufio.NewScanner(conn)
	fmt.Println("INPUT", input)
	for input.Scan() {
		// Split que separa por espaço a String que foi digitado pelo usuário
		args := strings.Split(input.Text(), " ")
		fmt.Println("For:", args)
		comando := args[0]

		switch comando {
		case "/nick":
			nick := strings.Split(input.Text(), " ")
			if len(nick) > 1 {
				messages <- "O " + apelido + " mudou seu nome para " + args[1]
				delete(uprivate, apelido)
				fmt.Println(args)
				apelido = args[1]
				uprivate[apelido] = ch
				ch <- "Seu novo nick é: " + apelido
			}

		case "/shownick":
			fmt.Println(apelido)
			ch <- "Nick: " + apelido

		case "/exit":
			leaving <- ch
			messages <- apelido + " se foi "
			conn.Close()

		case "/private":
			private <- apelido + " " + input.Text()

		case "/users":
			lista := "Clientes Online: "
			for k, _ := range uprivate {
				lista += k + ", "
			}
			ch <- lista

		case "/bot":
			texto := strings.SplitAfterN(input.Text(), " ", 2)
			messages <- "Bot: " + reverse_string(texto[1])

		default:
			messages <- apelido + ":" + input.Text()
		}
	}
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
