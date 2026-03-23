import Foundation
import Combine

// MARK: - Data Types

struct GameItem {
    let id: String
    let name: String
    let description: String
    let takeable: Bool
    let readText: String?
}

struct GameExit {
    let direction: String
    let roomID: String
    var locked: Bool
    let keyID: String?
}

struct GameRoom {
    let id: String
    let name: String
    let description: String
    let darkDesc: String?
    var exits: [String: GameExit]
    var items: [String: GameItem]
    let dark: Bool
}

// MARK: - Game Engine

@MainActor
final class GameEngine: ObservableObject {

    @Published private(set) var messages: [String] = []
    @Published private(set) var isRunning: Bool = true
    @Published private(set) var hasWon: Bool = false

    private var rooms: [String: GameRoom] = [:]
    private var playerRoomID: String = "bedroom"
    private var inventory: [String: GameItem] = [:]

    init() {
        setupWorld()
        emitIntro()
        cmdLook()
    }

    // MARK: - Setup

    private func setupWorld() {
        rooms = [
            "bedroom": GameRoom(
                id: "bedroom",
                name: "Dusty Bedroom",
                description: "A musty bedroom with a sagging four-poster bed. Morning light filters "
                    + "through grimy windows, casting pale rectangles on the dusty floorboards. "
                    + "A narrow door leads south to the hallway.",
                darkDesc: nil,
                exits: [
                    "s": GameExit(direction: "s", roomID: "hallway", locked: false, keyID: nil)
                ],
                items: [
                    "matches": GameItem(
                        id: "matches",
                        name: "box of matches",
                        description: "A small box of wooden matches. Strike one to light things.",
                        takeable: true,
                        readText: nil
                    )
                ],
                dark: false
            ),

            "hallway": GameRoom(
                id: "hallway",
                name: "Grand Hallway",
                description: "A long hallway with faded portraits on panelled walls. Dust motes drift "
                    + "in the stale air. Doors lead north to the bedroom, east to the library, "
                    + "and west to the kitchen. A heavy oak door to the south bears a brass lock.",
                darkDesc: nil,
                exits: [
                    "n": GameExit(direction: "n", roomID: "bedroom",  locked: false, keyID: nil),
                    "e": GameExit(direction: "e", roomID: "library",  locked: false, keyID: nil),
                    "w": GameExit(direction: "w", roomID: "kitchen",  locked: false, keyID: nil),
                    "s": GameExit(direction: "s", roomID: "garden",   locked: true,  keyID: "brass_key")
                ],
                items: [
                    "portrait": GameItem(
                        id: "portrait",
                        name: "faded portrait",
                        description: "A portrait of a stern Victorian gentleman in a dark coat. "
                            + "His painted eyes seem to follow you with unsettling intensity.",
                        takeable: false,
                        readText: nil
                    )
                ],
                dark: false
            ),

            "library": GameRoom(
                id: "library",
                name: "Manor Library",
                description: "Floor-to-ceiling bookshelves line every wall, packed with ancient volumes. "
                    + "A mahogany reading desk sits beneath a dusty window. An unlit candle "
                    + "and a worn tome rest on its surface. The hallway lies to the west.",
                darkDesc: nil,
                exits: [
                    "w": GameExit(direction: "w", roomID: "hallway", locked: false, keyID: nil)
                ],
                items: [
                    "candle": GameItem(
                        id: "candle",
                        name: "white candle",
                        description: "A tall white candle set in a brass holder. The wick is unburnt — it needs to be lit.",
                        takeable: true,
                        readText: nil
                    ),
                    "old_tome": GameItem(
                        id: "old_tome",
                        name: "old tome",
                        description: "A large dusty tome titled 'A History of Thornwood Manor'. "
                            + "It looks like it might contain useful information.",
                        takeable: true,
                        readText: """
You carefully turn the yellowed pages, breathing in the scent of old paper.
Near the back, a handwritten note in cramped script reads:

  "Should you find yourself locked within these walls, seek the
   brass key hidden where provisions are stored, away from the light.
   Bring fire to pierce the darkness, and you shall find your way."

Below, someone has scrawled in a different hand: "-- D.T., 1887"
"""
                    )
                ],
                dark: false
            ),

            "kitchen": GameRoom(
                id: "kitchen",
                name: "Manor Kitchen",
                description: "A cavernous kitchen with a cold stone fireplace and a heavy iron range. "
                    + "Copper pots hang from ceiling hooks. The room smells faintly of ash and old grease. "
                    + "A door to the north leads to the pantry; the hallway is to the east.",
                darkDesc: nil,
                exits: [
                    "e": GameExit(direction: "e", roomID: "hallway", locked: false, keyID: nil),
                    "n": GameExit(direction: "n", roomID: "pantry",  locked: false, keyID: nil)
                ],
                items: [:],
                dark: false
            ),

            "pantry": GameRoom(
                id: "pantry",
                name: "Dark Pantry",
                description: "A small windowless pantry. Rows of dusty jars and old provisions line "
                    + "the wooden shelves. Something catches the candlelight in the far corner.",
                darkDesc: "A small windowless pantry, pitch black without a light source. "
                    + "You can feel the shelves around you but can't make out anything clearly.",
                exits: [
                    "s": GameExit(direction: "s", roomID: "kitchen", locked: false, keyID: nil)
                ],
                items: [
                    "brass_key": GameItem(
                        id: "brass_key",
                        name: "brass key",
                        description: "A small brass key decorated with an engraved garden motif. It would fit a garden gate.",
                        takeable: true,
                        readText: nil
                    )
                ],
                dark: true
            ),

            "garden": GameRoom(
                id: "garden",
                name: "Manor Garden",
                description: "You step into a beautiful overgrown garden. Ancient oak trees arch overhead, "
                    + "dappled sunlight filtering through their rustling leaves. "
                    + "Roses run wild along crumbling stone walls, and the air smells of earth and freedom. "
                    + "The garden gate stands open before you.",
                darkDesc: nil,
                exits: [
                    "n": GameExit(direction: "n", roomID: "hallway", locked: false, keyID: nil)
                ],
                items: [:],
                dark: false
            )
        ]
    }

    private func emitIntro() {
        emit("╔══════════════════════════════════════════════════╗")
        emit("║         ESCAPE FROM THORNWOOD MANOR              ║")
        emit("╚══════════════════════════════════════════════════╝")
        emit("")
        emit("You wake with a start. The last thing you remember is exploring")
        emit("the abandoned Thornwood Manor on a dare when a heavy door slammed")
        emit("shut behind you. The air is stale and the house is silent.")
        emit("You must find a way to escape.")
        emit("")
        emit("Type 'help' for a list of commands.")
    }

    // MARK: - Output

    private func emit(_ text: String) {
        messages.append(text)
    }

    func addMessage(_ text: String) {
        messages.append(text)
    }

    // MARK: - Helpers

    private var currentRoom: GameRoom { rooms[playerRoomID]! }

    private var hasLight: Bool { inventory["lit_candle"] != nil }

    private func itemMatches(id: String, name: String, query: String) -> Bool {
        let q = query.trimmingCharacters(in: .whitespaces).lowercased()
        guard !q.isEmpty else { return false }
        let qNorm = q.replacingOccurrences(of: " ", with: "_")
        let idL   = id.lowercased()
        let nameL = name.lowercased()
        return idL == qNorm || nameL == q || nameL.contains(q) || idL.contains(qNorm)
    }

    private func dirLabel(_ dir: String) -> String {
        switch dir {
        case "n": return "north"
        case "s": return "south"
        case "e": return "east"
        case "w": return "west"
        default:  return dir
        }
    }

    private func normalizeDir(_ s: String) -> String {
        switch s.lowercased() {
        case "north", "n": return "n"
        case "south", "s": return "s"
        case "east",  "e": return "e"
        case "west",  "w": return "w"
        default:           return s.lowercased()
        }
    }

    /// Searches inventory first, then the current room (blocked in darkness).
    private func findItem(query: String) -> GameItem? {
        if let item = inventory.first(where: { itemMatches(id: $0.key, name: $0.value.name, query: query) })?.value {
            return item
        }
        let room = currentRoom
        guard !room.dark || hasLight else { return nil }
        return room.items.first(where: { itemMatches(id: $0.key, name: $0.value.name, query: query) })?.value
    }

    // MARK: - Commands

    private func cmdLook() {
        let room = currentRoom
        let isDark = room.dark && !hasLight

        emit("")
        emit(room.name)
        emit(String(repeating: "─", count: room.name.count))

        if isDark, let darkDesc = room.darkDesc {
            emit(darkDesc)
        } else {
            emit(room.description)
        }

        if !isDark && !room.items.isEmpty {
            let names = room.items.values.map(\.name).sorted()
            emit("")
            emit("You can see: \(names.joined(separator: ", ")).")
        }

        let dirs = room.exits.keys.sorted()
        let parts = dirs.map { d -> String in
            var label = dirLabel(d)
            if room.exits[d]!.locked { label += " (locked)" }
            return label
        }
        emit("Exits: \(parts.joined(separator: ", ")).")
    }

    private func cmdGo(dir: String) {
        let direction = normalizeDir(dir)

        guard var exit = currentRoom.exits[direction] else {
            emit("You can't go that way.")
            return
        }

        if exit.locked {
            guard let keyID = exit.keyID else {
                emit("The way is blocked.")
                return
            }
            guard let key = inventory[keyID] else {
                emit("The door is locked. You'll need a key to open it.")
                return
            }
            exit.locked = false
            rooms[playerRoomID]?.exits[direction] = exit
            emit("You use the \(key.name) to unlock the door.")
        }

        playerRoomID = exit.roomID
        cmdLook()

        if playerRoomID == "garden" {
            emit("")
            emit("╔══════════════════════════════════════════════════╗")
            emit("║  Congratulations! You have escaped Thornwood Manor!  ║")
            emit("╚══════════════════════════════════════════════════╝")
            emit("")
            emit("You stride through the garden gate into warm afternoon sunlight.")
            emit("The old manor looms behind you, but you are free at last.")
            emit("")
            emit("Thanks for playing!")
            hasWon = true
            isRunning = false
        }
    }

    private func cmdTake(query: String) {
        let room = currentRoom
        guard !room.dark || hasLight else {
            emit("It's too dark to find anything. You need a light source.")
            return
        }
        guard let (foundID, found) = room.items.first(where: { itemMatches(id: $0.key, name: $0.value.name, query: query) }) else {
            emit("You don't see '\(query)' here.")
            return
        }
        guard found.takeable else {
            emit("You can't take the \(found.name).")
            return
        }
        inventory[foundID] = found
        rooms[playerRoomID]?.items.removeValue(forKey: foundID)
        emit("You pick up the \(found.name).")
    }

    private func cmdDrop(query: String) {
        guard let (foundID, found) = inventory.first(where: { itemMatches(id: $0.key, name: $0.value.name, query: query) }) else {
            emit("You're not carrying '\(query)'.")
            return
        }
        inventory.removeValue(forKey: foundID)
        rooms[playerRoomID]?.items[foundID] = found
        emit("You set down the \(found.name).")
    }

    private func cmdInventory() {
        guard !inventory.isEmpty else {
            emit("You are not carrying anything.")
            return
        }
        emit("You are carrying:")
        for name in inventory.values.map(\.name).sorted() {
            emit("  - \(name)")
        }
    }

    private func cmdExamine(query: String) {
        guard let item = findItem(query: query) else {
            emit(currentRoom.dark && !hasLight
                 ? "It's too dark to examine anything."
                 : "You don't see '\(query)' here.")
            return
        }
        emit(item.description)
    }

    private func cmdRead(query: String) {
        guard let item = findItem(query: query) else {
            emit(currentRoom.dark && !hasLight
                 ? "It's too dark to read anything."
                 : "You don't see '\(query)' to read.")
            return
        }
        guard let text = item.readText else {
            emit("There's nothing to read on the \(item.name).")
            return
        }
        emit(text)
    }

    private func lightCandle() {
        inventory.removeValue(forKey: "candle")
        inventory.removeValue(forKey: "matches")
        inventory["lit_candle"] = GameItem(
            id: "lit_candle",
            name: "lit candle",
            description: "A white candle burning with a steady flame. Warm golden light pushes back the darkness.",
            takeable: true,
            readText: nil
        )
        emit("You strike a match and touch it to the wick. The candle catches,")
        emit("and warm golden light blooms around you.")
        if currentRoom.dark {
            emit("")
            cmdLook()
        }
    }

    private func cmdUse(itemQuery: String, targetQuery: String) {
        guard let (useID, useItem) = inventory.first(where: { itemMatches(id: $0.key, name: $0.value.name, query: itemQuery) }) else {
            emit("You're not carrying '\(itemQuery)'.")
            return
        }

        switch useID {
        case "matches":
            if inventory["candle"] != nil {
                lightCandle()
            } else {
                emit("You strike a match. The flame sputters out. Nothing here to light.")
            }

        case "candle":
            if inventory["matches"] != nil {
                lightCandle()
            } else {
                emit("The candle is unlit. You need something to light it with.")
            }

        case "lit_candle":
            emit("The candle flame dances. It lights the area around you.")

        case "brass_key":
            let room = currentRoom
            if let dir = room.exits.first(where: { $0.value.locked && $0.value.keyID == "brass_key" })?.key {
                rooms[playerRoomID]?.exits[dir]?.locked = false
                emit("You unlock the door to the \(dirLabel(dir)) with the brass key.")
            } else {
                emit("There's no lock here for the brass key.")
            }

        case "old_tome":
            cmdRead(query: "old tome")

        default:
            emit("You're not sure how to use the \(useItem.name).")
        }
    }

    private func showHelp() {
        emit("""

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

""")
    }

    // MARK: - Public Input Handler

    func handleInput(_ line: String) {
        let trimmed = line.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }

        let parts = trimmed.split(separator: " ", omittingEmptySubsequences: true).map(String.init)
        let verb = parts[0].lowercased()
        let rest = parts.dropFirst().joined(separator: " ")

        switch verb {
        case "look", "l":
            if rest.isEmpty {
                cmdLook()
            } else {
                let arg = rest.lowercased().hasPrefix("at ") ? String(rest.dropFirst(3)) : rest
                cmdExamine(query: arg)
            }

        case "go":
            rest.isEmpty ? emit("Go where? (north, south, east, west)") : cmdGo(dir: rest)

        case "n", "north", "s", "south", "e", "east", "w", "west":
            cmdGo(dir: verb)

        case "take", "get":
            rest.isEmpty ? emit("Take what?") : cmdTake(query: rest)

        case "pick":
            if rest.lowercased().hasPrefix("up ") {
                cmdTake(query: String(rest.dropFirst(3)))
            } else if !rest.isEmpty {
                cmdTake(query: rest)
            } else {
                emit("Pick up what?")
            }

        case "drop":
            rest.isEmpty ? emit("Drop what?") : cmdDrop(query: rest)

        case "inventory", "inv", "i":
            cmdInventory()

        case "examine", "x", "inspect":
            rest.isEmpty ? cmdLook() : cmdExamine(query: rest)

        case "read":
            rest.isEmpty ? emit("Read what?") : cmdRead(query: rest)

        case "use":
            if rest.isEmpty {
                emit("Use what?")
            } else {
                let lower = rest.lowercased()
                let itemQ: String
                let targetQ: String
                if let range = lower.range(of: " on ") {
                    let offset = lower.distance(from: lower.startIndex, to: range.lowerBound)
                    itemQ   = String(rest.prefix(offset))
                    targetQ = String(rest.suffix(from: rest.index(rest.startIndex, offsetBy: offset + 4)))
                } else if let range = lower.range(of: " with ") {
                    let offset = lower.distance(from: lower.startIndex, to: range.lowerBound)
                    itemQ   = String(rest.prefix(offset))
                    targetQ = String(rest.suffix(from: rest.index(rest.startIndex, offsetBy: offset + 6)))
                } else {
                    itemQ   = rest
                    targetQ = ""
                }
                cmdUse(itemQuery: itemQ, targetQuery: targetQ)
            }

        case "help", "h", "?":
            showHelp()

        case "quit", "exit", "q":
            emit("Thanks for playing. Goodbye!")
            isRunning = false

        default:
            emit("Unknown command '\(verb)'. Type 'help' for a list of commands.")
        }
    }
}
