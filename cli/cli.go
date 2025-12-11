package cli

import (
	"bufio"
	"fmt"
	"io"
	"jsondb/db"
	helper "jsondb/helpers"
	"os"
	"strings"
)

var PREFIX = "> "

type CLI struct {
	db *db.DB
}

func New() *CLI { return &CLI{db: nil} }

func (c *CLI) ensureDBLoaded() bool { return c.db != nil }

func (c *CLI) Run() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("enter commands (type 'help' for list, 'quit' to exit)")

	for {
		fmt.Print(PREFIX)
		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF (Ctrl+C) or other read error
			if err == io.EOF {
				fmt.Print("\njsondb: bye\n\n")
				return
			}
			c.error(fmt.Sprintf("read error: %v\n", err))
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "help", "h", "?":
			c.displayHelp()
		case "quit", "exit", "q":
			fmt.Print("\njsondb: bye\n\n")
			return
		case "info":
			c.info()
		case "init":
			if err := c.initDB(args); err != nil {
				c.error(err.Error())
				continue
			}
			PREFIX = fmt.Sprintf(" (%s)> ", args[0])
		case "load":
			if err := c.loadDB(args); err != nil {
				c.error(err.Error())
				continue
			}
			PREFIX = fmt.Sprintf(" (%s)> ", args[0])
		case "create":
			if err := c.createCollection(args); err != nil {
				c.error(err.Error())
			}
		default:
			c.error("incorrect command")
		}
	}
}

func (c *CLI) displayHelp() {
	fmt.Println()
	fmt.Println("Usage: <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init           					   - create DB directory if it doesn't exist")
	fmt.Println("  load                                - load the DB (directory must exist)")
	fmt.Println("  info							       - display info about the loaded database")
	fmt.Println("  create <collection>                 - Create a new collection in the database")
	fmt.Println("  insert <collection> <data>     	   - Insert a new record into the database")
	fmt.Println("  query  <collection> <clause>        - Query records from the database")
	fmt.Println("  update <collection> <clause> <data> - Update records in the database")
	fmt.Println("  delete <collection> <clause>        - Delete records from the database")
	fmt.Println("  drop   <collection>   		       - Drop a collection from the database")
	fmt.Println("  help        	                       - Display this help message")
	fmt.Println()
}

func (c *CLI) info() {
	fmt.Println("Info:")
	if c.ensureDBLoaded() {
		fmt.Printf("  Database loaded: %s\n", c.db.BaseDir)
	} else {
		fmt.Println("  No database loaded")
	}
	fmt.Println()
}

func (c *CLI) error(message string) {
	fmt.Println()
	fmt.Println("jsondb:", message)
	fmt.Println()
}

func (c *CLI) errorFatal(message string) {
	fmt.Println()
	fmt.Println("jsondb:", message)
	fmt.Println()
	os.Exit(1)
}

func (c *CLI) initDB(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("init command requires exactly one argument: the database file name")
	}

	c.db = db.New(args[0])

	if err := c.db.Init(); err != nil {
		return fmt.Errorf("init command failed: %v", err)
	}
	return nil
}

func (c *CLI) loadDB(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("load command requires name of an existing database")
	}

	path := args[0]
	exists, _, _ := helper.PathExist(path)
	if !exists {
		return fmt.Errorf("database with path %s does not exist", path)
	}

	c.db = db.New(path)
	return nil
}

func (c *CLI) createCollection(args []string) error {
	if !c.ensureDBLoaded() {
		return fmt.Errorf("no database loaded. Use 'init' or 'load' command first")
	}
	if len(args) != 1 {
		return fmt.Errorf("create command requires exactly one argument: the collection name")
	}

	return c.db.CreateCollection(args[0])
}
