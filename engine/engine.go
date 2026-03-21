package engine

import (
	"sort"
	"strings"
)

// ─── Data Types ────────────────────────────────────────────────────────────

// Item is a single object in the world. All fields are exported for serialization.
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Takeable    bool   `json:"takeable"`
	ReadText    string `json:"readText,omitempty"`
}

// Exit connects a room to another room in one direction.
type Exit struct {
	Direction string `json:"direction"`
	RoomID    string `json:"roomID"`
	Locked    bool   `json:"locked"`
	KeyID     string `json:"keyID,omitempty"`
}

// Room is a location in the world.
type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	DarkDesc    string            `json:"darkDesc,omitempty"`
	Exits       map[string]Exit   `json:"exits"`
	Items       map[string]Item   `json:"items"`
	Dark        bool              `json:"dark,omitempty"`
}

// ─── Game Session ──────────────────────────────────────────────────────────

// GameSession holds all mutable game state. It is safe to serialize with
// encoding/json — no pointers, no unexported fields that matter for state.
type GameSession struct {
	PlayerRoomID string            `json:"playerRoomID"`
	Inventory    map[string]Item   `json:"inventory"`
	Rooms        map[string]Room   `json:"rooms"`
	Won          bool              `json:"won"`
	Running      bool              `json:"running"`
	TurnCount    int               `json:"turnCount"`

	// output is transient I/O state; excluded from serialization.
	output []string `json:"-"`
}

// NewSession creates and returns a fully initialised game session.
func NewSession() *GameSession {
	s := &GameSession{
		PlayerRoomID: "bedroom",
		Inventory:    map[string]Item{},
		Running:      true,
		Rooms: map[string]Room{
			"bedroom": {
				ID:   "bedroom",
				Name: "Dusty Bedroom",
				Description: "A musty bedroom with a sagging four-poster bed. Morning light filters " +
					"through grimy windows, casting pale rectangles on the dusty floorboards. " +
					"A narrow door leads south to the hallway.",
				Exits: map[string]Exit{
					"s": {Direction: "s", RoomID: "hallway"},
				},
				Items: map[string]Item{
					"matches": {
						ID:          "matches",
						Name:        "box of matches",
						Description: "A small box of wooden matches. Strike one to light things.",
						Takeable:    true,
					},
				},
			},
			"hallway": {
				ID:   "hallway",
				Name: "Grand Hallway",
				Description: "A long hallway with faded portraits on panelled walls. Dust motes drift " +
					"in the stale air. Doors lead north to the bedroom, east to the library, " +
					"and west to the kitchen. A heavy oak door to the south bears a brass lock.",
				Exits: map[string]Exit{
					"n": {Direction: "n", RoomID: "bedroom"},
					"e": {Direction: "e", RoomID: "library"},
					"w": {Direction: "w", RoomID: "kitchen"},
					"s": {Direction: "s", RoomID: "garden", Locked: true, KeyID: "brass_key"},
				},
				Items: map[string]Item{
					"portrait": {
						ID:   "portrait",
						Name: "faded portrait",
						Description: "A portrait of a stern Victorian gentleman in a dark coat. " +
							"His painted eyes seem to follow you with unsettling intensity.",
						Takeable: false,
					},
				},
			},
			"library": {
				ID:   "library",
				Name: "Manor Library",
				Description: "Floor-to-ceiling bookshelves line every wall, packed with ancient volumes. " +
					"A mahogany reading desk sits beneath a dusty window. An unlit candle " +
					"and a worn tome rest on its surface. The hallway lies to the west.",
				Exits: map[string]Exit{
					"w": {Direction: "w", RoomID: "hallway"},
				},
				Items: map[string]Item{
					"candle": {
						ID:          "candle",
						Name:        "white candle",
						Description: "A tall white candle set in a brass holder. The wick is unburnt — it needs to be lit.",
						Takeable:    true,
					},
					"old_tome": {
						ID:   "old_tome",
						Name: "old tome",
						Description: "A large dusty tome titled 'A History of Thornwood Manor'. " +
							"It looks like it might contain useful information.",
						Takeable: true,
						ReadText: `You carefully turn the yellowed pages, breathing in the scent of old paper.
Near the back, a handwritten note in cramped script reads:

  "Should you find yourself locked within these walls, seek the
   brass key hidden where provisions are stored, away from the light.
   Bring fire to pierce the darkness, and you shall find your way."

Below, someone has scrawled in a different hand: "-- D.T., 1887"`,
					},
				},
			},
			"kitchen": {
				ID:   "kitchen",
				Name: "Manor Kitchen",
				Description: "A cavernous kitchen with a cold stone fireplace and a heavy iron range. " +
					"Copper pots hang from ceiling hooks. The room smells faintly of ash and old grease. " +
					"A door to the north leads to the pantry; the hallway is to the east.",
				Exits: map[string]Exit{
					"e": {Direction: "e", RoomID: "hallway"},
					"n": {Direction: "n", RoomID: "pantry"},
				},
				Items: map[string]Item{},
			},
			"pantry": {
				ID:   "pantry",
				Name: "Dark Pantry",
				Description: "A small windowless pantry. Rows of dusty jars and old provisions line " +
					"the wooden shelves. Something catches the candlelight in the far corner.",
				DarkDesc: "A small windowless pantry, pitch black without a light source. " +
					"You can feel the shelves around you but can't make out anything clearly.",
				Exits: map[string]Exit{
					"s": {Direction: "s", RoomID: "kitchen"},
				},
				Items: map[string]Item{
					"brass_key": {
						ID:          "brass_key",
						Name:        "brass key",
						Description: "A small brass key decorated with an engraved garden motif. It would fit a garden gate.",
						Takeable:    true,
					},
				},
				Dark: true,
			},
			"garden": {
				ID:   "garden",
				Name: "Manor Garden",
				Description: "You step into a beautiful overgrown garden. Ancient oak trees arch overhead, " +
					"dappled sunlight filtering through their rustling leaves. " +
					"Roses run wild along crumbling stone walls, and the air smells of earth and freedom. " +
					"The garden gate stands open before you.",
				Exits: map[string]Exit{
					"n": {Direction: "n", RoomID: "hallway"},
				},
				Items: map[string]Item{},
			},
		},
	}

	s.emit("╔══════════════════════════════════════════════════╗")
	s.emit("║         ESCAPE FROM THORNWOOD MANOR              ║")
	s.emit("╚══════════════════════════════════════════════════╝")
	s.emit("")
	s.emit("You wake with a start. The last thing you remember is exploring")
	s.emit("the abandoned Thornwood Manor on a dare when a heavy door slammed")
	s.emit("shut behind you. The air is stale and the house is silent.")
	s.emit("You must find a way to escape.")
	s.emit("")
	s.emit("Type 'help' for a list of commands.")
	s.cmdLook()

	return s
}

// ─── Output ────────────────────────────────────────────────────────────────

func (s *GameSession) emit(text string) {
	s.output = append(s.output, text)
}

// FlushOutput returns all buffered output lines and clears the buffer.
func (s *GameSession) FlushOutput() []string {
	out := s.output
	s.output = nil
	return out
}

// IsRunning reports whether the game loop should continue.
func (s *GameSession) IsRunning() bool {
	return s.Running
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func (s *GameSession) currentRoom() Room {
	return s.Rooms[s.PlayerRoomID]
}

func (s *GameSession) hasLight() bool {
	_, ok := s.Inventory["lit_candle"]
	return ok
}

// itemMatches returns true if query matches an item's ID or name (exact or partial).
func itemMatches(id, name, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return false
	}
	qNorm := strings.ReplaceAll(q, " ", "_")
	idL := strings.ToLower(id)
	nameL := strings.ToLower(name)
	return idL == qNorm ||
		nameL == q ||
		strings.Contains(nameL, q) ||
		strings.Contains(idL, qNorm)
}

func dirLabel(dir string) string {
	switch dir {
	case "n":
		return "north"
	case "s":
		return "south"
	case "e":
		return "east"
	case "w":
		return "west"
	default:
		return dir
	}
}

func normalizeDir(d string) string {
	switch strings.ToLower(d) {
	case "north", "n":
		return "n"
	case "south", "s":
		return "s"
	case "east", "e":
		return "e"
	case "west", "w":
		return "w"
	default:
		return strings.ToLower(d)
	}
}

// findItem searches inventory first, then the current room (respecting darkness).
// Returns the item and true if found, or zero value and false if not.
func (s *GameSession) findItem(query string) (Item, bool) {
	for id, item := range s.Inventory {
		if itemMatches(id, item.Name, query) {
			return item, true
		}
	}
	room := s.currentRoom()
	if !room.Dark || s.hasLight() {
		for id, item := range room.Items {
			if itemMatches(id, item.Name, query) {
				return item, true
			}
		}
	}
	return Item{}, false
}

// ─── Commands ──────────────────────────────────────────────────────────────

func (s *GameSession) cmdLook() {
	room := s.currentRoom()
	isDark := room.Dark && !s.hasLight()

	s.emit("")
	s.emit(room.Name)
	s.emit(strings.Repeat("─", len(room.Name)))

	if isDark && room.DarkDesc != "" {
		s.emit(room.DarkDesc)
	} else {
		s.emit(room.Description)
	}

	if !isDark && len(room.Items) > 0 {
		names := make([]string, 0, len(room.Items))
		for _, item := range room.Items {
			names = append(names, item.Name)
		}
		sort.Strings(names)
		s.emit("")
		s.emit("You can see: " + strings.Join(names, ", ") + ".")
	}

	dirs := make([]string, 0, len(room.Exits))
	for d := range room.Exits {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	parts := make([]string, 0, len(dirs))
	for _, d := range dirs {
		exit := room.Exits[d]
		label := dirLabel(d)
		if exit.Locked {
			label += " (locked)"
		}
		parts = append(parts, label)
	}
	s.emit("Exits: " + strings.Join(parts, ", ") + ".")
}

func (s *GameSession) cmdGo(dir string) {
	dir = normalizeDir(dir)
	room := s.currentRoom()

	exit, ok := room.Exits[dir]
	if !ok {
		s.emit("You can't go that way.")
		return
	}

	if exit.Locked {
		if exit.KeyID == "" {
			s.emit("The way is blocked.")
			return
		}
		key, hasKey := s.Inventory[exit.KeyID]
		if !hasKey {
			s.emit("The door is locked. You'll need a key to open it.")
			return
		}
		exit.Locked = false
		room.Exits[dir] = exit // write value back into the map
		s.emit("You use the " + key.Name + " to unlock the door.")
	}

	s.PlayerRoomID = exit.RoomID
	s.cmdLook()

	if s.PlayerRoomID == "garden" {
		s.emit("")
		s.emit("╔══════════════════════════════════════════════════╗")
		s.emit("║  Congratulations! You have escaped Thornwood Manor!  ║")
		s.emit("╚══════════════════════════════════════════════════╝")
		s.emit("")
		s.emit("You stride through the garden gate into warm afternoon sunlight.")
		s.emit("The old manor looms behind you, but you are free at last.")
		s.emit("")
		s.emit("Thanks for playing!")
		s.Won = true
		s.Running = false
	}
}

func (s *GameSession) cmdTake(query string) {
	room := s.currentRoom()
	if room.Dark && !s.hasLight() {
		s.emit("It's too dark to find anything. You need a light source.")
		return
	}

	var foundID string
	var found Item
	for id, item := range room.Items {
		if itemMatches(id, item.Name, query) {
			foundID = id
			found = item
			break
		}
	}

	if foundID == "" {
		s.emit("You don't see '" + query + "' here.")
		return
	}
	if !found.Takeable {
		s.emit("You can't take the " + found.Name + ".")
		return
	}

	s.Inventory[foundID] = found
	delete(room.Items, foundID) // room.Items is a map — deletion affects the original
	s.emit("You pick up the " + found.Name + ".")
}

func (s *GameSession) cmdDrop(query string) {
	room := s.currentRoom()

	var foundID string
	var found Item
	for id, item := range s.Inventory {
		if itemMatches(id, item.Name, query) {
			foundID = id
			found = item
			break
		}
	}

	if foundID == "" {
		s.emit("You're not carrying '" + query + "'.")
		return
	}

	delete(s.Inventory, foundID)
	room.Items[foundID] = found // room.Items is a map — assignment affects the original
	s.emit("You set down the " + found.Name + ".")
}

func (s *GameSession) cmdInventory() {
	if len(s.Inventory) == 0 {
		s.emit("You are not carrying anything.")
		return
	}
	s.emit("You are carrying:")
	names := make([]string, 0, len(s.Inventory))
	for _, item := range s.Inventory {
		names = append(names, "  - "+item.Name)
	}
	sort.Strings(names)
	for _, n := range names {
		s.emit(n)
	}
}

func (s *GameSession) cmdExamine(query string) {
	item, ok := s.findItem(query)
	if !ok {
		if s.currentRoom().Dark && !s.hasLight() {
			s.emit("It's too dark to examine anything.")
		} else {
			s.emit("You don't see '" + query + "' here.")
		}
		return
	}
	s.emit(item.Description)
}

func (s *GameSession) cmdRead(query string) {
	item, ok := s.findItem(query)
	if !ok {
		if s.currentRoom().Dark && !s.hasLight() {
			s.emit("It's too dark to read anything.")
		} else {
			s.emit("You don't see '" + query + "' to read.")
		}
		return
	}
	if item.ReadText == "" {
		s.emit("There's nothing to read on the " + item.Name + ".")
		return
	}
	s.emit(item.ReadText)
}

func (s *GameSession) lightCandle() {
	delete(s.Inventory, "candle")
	delete(s.Inventory, "matches")
	s.Inventory["lit_candle"] = Item{
		ID:          "lit_candle",
		Name:        "lit candle",
		Description: "A white candle burning with a steady flame. Warm golden light pushes back the darkness.",
		Takeable:    true,
	}
	s.emit("You strike a match and touch it to the wick. The candle catches,")
	s.emit("and warm golden light blooms around you.")

	if s.currentRoom().Dark {
		s.emit("")
		s.cmdLook()
	}
}

func (s *GameSession) cmdUse(itemQuery, targetQuery string) {
	var useID string
	var useItem Item
	for id, item := range s.Inventory {
		if itemMatches(id, item.Name, itemQuery) {
			useID = id
			useItem = item
			break
		}
	}

	if useID == "" {
		s.emit("You're not carrying '" + itemQuery + "'.")
		return
	}

	switch useID {
	case "matches":
		if _, hasCandle := s.Inventory["candle"]; hasCandle {
			s.lightCandle()
		} else {
			s.emit("You strike a match. The flame sputters out. Nothing here to light.")
		}

	case "candle":
		if _, hasMatches := s.Inventory["matches"]; hasMatches {
			s.lightCandle()
		} else {
			s.emit("The candle is unlit. You need something to light it with.")
		}

	case "lit_candle":
		s.emit("The candle flame dances. It lights the area around you.")

	case "brass_key":
		room := s.currentRoom()
		for dir, exit := range room.Exits {
			if exit.Locked && exit.KeyID == "brass_key" {
				exit.Locked = false
				room.Exits[dir] = exit // write value back into the map
				s.emit("You unlock the door to the " + dirLabel(dir) + " with the brass key.")
				return
			}
		}
		s.emit("There's no lock here for the brass key.")

	case "old_tome":
		s.cmdRead("old tome")

	default:
		s.emit("You're not sure how to use the " + useItem.Name + ".")
	}
}

func (s *GameSession) showHelp() {
	s.emit(`
Commands:
  look / l                      Describe your surroundings
  look at <item>                Examine something closely
  go <direction>                Move north, south, east, or west
  n / s / e / w                 Shorthand for directions
  take <item>  /  get <item>    Pick up an item
  drop <item>                   Drop an item from your inventory
  inventory / i                 List what you're carrying
  examine <item> / x <item>     Examine an item closely
  read <item>                   Read something
  use <item>                    Use an item
  use <item> on <target>        Use one item on another
  help / ?                      Show this help text
  quit / q                      Quit the game
`)
}

// ─── Public Input Handler ──────────────────────────────────────────────────

func (s *GameSession) HandleInput(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	s.TurnCount++

	parts := strings.Fields(line)
	verb := strings.ToLower(parts[0])
	rest := ""
	if len(parts) > 1 {
		rest = strings.Join(parts[1:], " ")
	}

	switch verb {
	case "look", "l":
		if rest == "" {
			s.cmdLook()
		} else {
			arg := rest
			if strings.HasPrefix(strings.ToLower(arg), "at ") {
				arg = arg[3:]
			}
			s.cmdExamine(arg)
		}

	case "go":
		if rest == "" {
			s.emit("Go where? (north, south, east, west)")
		} else {
			s.cmdGo(rest)
		}

	case "n", "north", "s", "south", "e", "east", "w", "west":
		s.cmdGo(verb)

	case "take", "get":
		if rest == "" {
			s.emit("Take what?")
		} else {
			s.cmdTake(rest)
		}

	case "pick":
		if strings.HasPrefix(strings.ToLower(rest), "up ") {
			s.cmdTake(rest[3:])
		} else if rest != "" {
			s.cmdTake(rest)
		} else {
			s.emit("Pick up what?")
		}

	case "drop":
		if rest == "" {
			s.emit("Drop what?")
		} else {
			s.cmdDrop(rest)
		}

	case "inventory", "inv", "i":
		s.cmdInventory()

	case "examine", "x", "inspect":
		if rest == "" {
			s.cmdLook()
		} else {
			s.cmdExamine(rest)
		}

	case "read":
		if rest == "" {
			s.emit("Read what?")
		} else {
			s.cmdRead(rest)
		}

	case "use":
		if rest == "" {
			s.emit("Use what?")
		} else {
			lower := strings.ToLower(rest)
			var itemQ, targetQ string
			if idx := strings.Index(lower, " on "); idx != -1 {
				itemQ = rest[:idx]
				targetQ = rest[idx+4:]
			} else if idx := strings.Index(lower, " with "); idx != -1 {
				itemQ = rest[:idx]
				targetQ = rest[idx+6:]
			} else {
				itemQ = rest
			}
			s.cmdUse(itemQ, targetQ)
		}

	case "help", "h", "?":
		s.showHelp()

	case "quit", "exit", "q":
		s.emit("Thanks for playing. Goodbye!")
		s.Running = false

	default:
		s.emit("Unknown command '" + verb + "'. Type 'help' for a list of commands.")
	}
}
