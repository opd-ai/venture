package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/ui"
)

func main() {
	// Command-line flags
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	width := flag.Int("width", 600, "Mailbox UI width")
	height := flag.Int("height", 400, "Mailbox UI height")
	mode := flag.String("mode", "inbox", "View mode (inbox, outbox, compose, detail)")
	output := flag.String("output", "", "Output PNG file (optional)")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	// Validate genre
	validGenres := map[string]bool{
		"fantasy": true, "scifi": true, "horror": true,
		"cyberpunk": true, "postapoc": true,
	}
	if !validGenres[*genreID] {
		fmt.Printf("Error: Invalid genre '%s'. Valid genres: fantasy, scifi, horror, cyberpunk, postapoc\n", *genreID)
		os.Exit(1)
	}

	// Create mailbox UI
	mailboxUI := ui.NewMailboxUI(0, 0, *width, *height, *genreID)

	// Create sample mail data
	mailComp := createSampleMailData()

	// Load data into UI
	mailboxUI.LoadFromMailComponent(mailComp)

	// Set view mode
	switch *mode {
	case "inbox":
		mailboxUI.SwitchView(ui.ViewInbox)
	case "outbox":
		mailboxUI.SwitchView(ui.ViewOutbox)
	case "compose":
		mailboxUI.SwitchView(ui.ViewCompose)
		// Pre-fill compose form for demo
		mailboxUI.ComposeRecipient = "player2"
		mailboxUI.ComposeSubject = "Greetings from the test tool"
		mailboxUI.ComposeBody = "This is a test message created by the mailbox UI test tool. It demonstrates the compose interface with text wrapping and attachment support."
		mailboxUI.AddAttachment(100)
		mailboxUI.AddAttachment(101)
	case "detail":
		mailboxUI.SwitchView(ui.ViewInbox)
		if len(mailboxUI.InboxMessages) > 0 {
			mailboxUI.OpenSelectedMessage()
		}
	default:
		fmt.Printf("Error: Invalid mode '%s'. Valid modes: inbox, outbox, compose, detail\n", *mode)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("=== Mailbox UI Test ===\n")
		fmt.Printf("Genre: %s\n", *genreID)
		fmt.Printf("Size: %dx%d\n", *width, *height)
		fmt.Printf("View Mode: %s\n", *mode)
		fmt.Printf("Inbox Messages: %d\n", len(mailboxUI.InboxMessages))
		fmt.Printf("Outbox Messages: %d\n", len(mailboxUI.OutboxMessages))
		fmt.Printf("Unread Count: %d\n", mailboxUI.GetUnreadCount())
		fmt.Println()
	}

	// Render UI
	img := mailboxUI.Render()

	// Save to file if output specified
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		if err := png.Encode(f, img); err != nil {
			fmt.Printf("Error encoding PNG: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Mailbox UI saved to %s\n", *output)
	} else {
		fmt.Println("Mailbox UI rendered successfully (use -output flag to save PNG)")
	}

	// Display navigation help
	if *verbose {
		fmt.Println("\n=== Navigation Methods ===")
		fmt.Println("SelectNext()      - Move to next message")
		fmt.Println("SelectPrevious()  - Move to previous message")
		fmt.Println("SwitchView(mode)  - Change view (ViewInbox, ViewOutbox, ViewCompose, ViewMessageDetail)")
		fmt.Println("OpenSelectedMessage() - View message details")
		fmt.Println("CloseMessageDetail()  - Return to list view")
		fmt.Println("AddAttachment(id)     - Add item to compose")
		fmt.Println("RemoveAttachment(idx) - Remove attachment")
		fmt.Println("ClearCompose()        - Reset compose form")
		fmt.Println()
	}

	// Display message statistics
	if *verbose {
		fmt.Println("=== Message Statistics ===")
		for i, msg := range mailboxUI.InboxMessages {
			fmt.Printf("[%d] From: %s | Subject: %s | Status: %s | Unread: %v\n",
				i, msg.SenderID, msg.Subject, msg.Status.String(), msg.IsUnread)
		}
		if len(mailboxUI.InboxMessages) == 0 {
			fmt.Println("(No inbox messages)")
		}
		fmt.Println()
	}

	// Test navigation if verbose
	if *verbose && *mode == "inbox" && len(mailboxUI.InboxMessages) > 1 {
		fmt.Println("=== Testing Navigation ===")
		mailboxUI.SelectNext()
		fmt.Printf("After SelectNext: SelectedInboxIndex = %d\n", mailboxUI.SelectedInboxIndex)
		mailboxUI.SelectPrevious()
		fmt.Printf("After SelectPrevious: SelectedInboxIndex = %d\n", mailboxUI.SelectedInboxIndex)
	}
}

// createSampleMailData creates sample mail data for testing
func createSampleMailData() *engine.MailComponent {
	mailComp := engine.NewMailComponent()
	now := time.Now().Unix()

	// Add inbox messages
	mailComp.Inbox = []*engine.MailMessage{
		{
			ID:          "msg001",
			SenderID:    "player2",
			RecipientID: "player1",
			Subject:     "Welcome to Venture!",
			Body:        "Hello adventurer! Welcome to the world of Venture. This is your first mail message. Explore, battle monsters, and discover the procedurally generated wonders that await!",
			Attachments: []uint64{100, 101},
			Postage:     12,
			SentAt:      now - 3600,
			DeliveredAt: now - 1800,
		},
		{
			ID:          "msg002",
			SenderID:    "npc_merchant",
			RecipientID: "player1",
			Subject:     "Special Offer: Legendary Sword",
			Body:        "Greetings! I have acquired a legendary sword that might interest you. Visit my shop in the town square for more details. Limited time offer!",
			Attachments: []uint64{},
			Postage:     10,
			SentAt:      now - 7200,
			DeliveredAt: now - 5400,
		},
		{
			ID:          "msg003",
			SenderID:    "guild_master",
			RecipientID: "player1",
			Subject:     "Quest Assignment: Dragon Slayer",
			Body:        "The guild has received reports of a dragon terrorizing nearby villages. Your skills are needed. Meet at the guild hall to receive the quest details and rewards.",
			Attachments: []uint64{200},
			Postage:     15,
			SentAt:      now - 10800,
			DeliveredAt: now - 9000,
		},
		{
			ID:          "msg004",
			SenderID:    "player3",
			RecipientID: "player1",
			Subject:     "Party Invitation",
			Body:        "Hey! A group of us are forming a party to explore the Cursed Dungeon. Want to join? Meet us at the tavern tonight.",
			Attachments: []uint64{},
			Postage:     10,
			SentAt:      now - 14400,
			DeliveredAt: now - 12600,
		},
		{
			ID:          "msg005",
			SenderID:    "blacksmith",
			RecipientID: "player1",
			Subject:     "Your Order is Ready",
			Body:        "The enchanted armor you ordered is complete. Come by the forge to collect it. Payment due on pickup: 500 gold.",
			Attachments: []uint64{300, 301},
			Postage:     10,
			SentAt:      now - 172800, // 2 days ago (should not be unread)
			DeliveredAt: now - 171000,
		},
	}

	// Add outbox messages
	mailComp.Outbox = []*engine.MailMessage{
		{
			ID:          "msg101",
			SenderID:    "player1",
			RecipientID: "player2",
			Subject:     "Thanks for the welcome!",
			Body:        "Thanks for the warm welcome message. I'm really enjoying Venture so far. Let's team up sometime!",
			Attachments: []uint64{},
			Postage:     10,
			SentAt:      now - 1800,
			DeliveredAt: now - 600, // Delivered
		},
		{
			ID:          "msg102",
			SenderID:    "player1",
			RecipientID: "player4",
			Subject:     "Item Trade Proposal",
			Body:        "I have a rare potion I'd like to trade for your magic staff. Let me know if you're interested.",
			Attachments: []uint64{400},
			Postage:     12,
			SentAt:      now - 900,
			DeliveredAt: 0, // In transit
		},
		{
			ID:          "msg103",
			SenderID:    "player1",
			RecipientID: "guild_master",
			Subject:     "Quest Status Update",
			Body:        "I've completed the first objective of the Dragon Slayer quest. The dragon has been driven from the northern village. Awaiting further instructions.",
			Attachments: []uint64{500, 501},
			Postage:     15,
			SentAt:      now - 300,
			DeliveredAt: 0, // In transit
		},
	}

	return mailComp
}
