package actions

import (
	//"encoding/json"
	//"fmt"
	//"time"

	"github.com/gobuffalo/buffalo"
	//"github.com/gorilla/websocket"
	//"github.com/gorilla/websocket"
)

type Person struct {
	Name string
	Age  int
}

// StatusHandler is a default handler to serve up
// a status page.
func StatusHandler(c buffalo.Context) error {

	/*
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("Client subscribed")

		myPerson := Person{
			Name: "Bill",
			Age:  0,
		}

		for {
			time.Sleep(2 * time.Second)
			if myPerson.Age < 40 {
				myJson, err := json.Marshal(myPerson)
				if err != nil {
					fmt.Println(err)
					return err
				}
				err = conn.WriteMessage(websocket.TextMessage, myJson)
				if err != nil {
					fmt.Println(err)
					break
				}
				myPerson.Age += 2
			} else {
				conn.Close()
				break
			}
		}
		fmt.Println("Client unsubscribed")
	*/

	return c.Render(200, r.HTML("status.html"))
}
