//
//  ContentView.swift
//  claude-text-adventure
//
//  Created by Stuart Dobbie on 22/3/2026.
//

import SwiftUI

struct ContentView: View {
    @StateObject private var engine = GameEngine()
    @State private var inputText = ""
    @FocusState private var inputFocused: Bool

    private let terminalGreen = Color(red: 0.2, green: 0.9, blue: 0.4)
    private let terminalDim   = Color(red: 0.15, green: 0.65, blue: 0.3)
    private let background    = Color(red: 0.05, green: 0.05, blue: 0.05)
    private let inputBg       = Color(red: 0.1, green: 0.1, blue: 0.1)

    var body: some View {
        ZStack {
            background.ignoresSafeArea()

            VStack(spacing: 0) {
                // ── Output log ───────────────────────────────────────
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 2) {
                            ForEach(Array(engine.messages.enumerated()), id: \.offset) { _, line in
                                Text(line.isEmpty ? " " : line)
                                    .font(.system(.footnote, design: .monospaced))
                                    .foregroundColor(terminalGreen)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                    .textSelection(.enabled)
                            }
                            // Invisible anchor at the bottom
                            Color.clear
                                .frame(height: 1)
                                .id("bottom")
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 8)
                    }
                    .onChange(of: engine.messages.count) {
                        withAnimation(.easeOut(duration: 0.15)) {
                            proxy.scrollTo("bottom", anchor: .bottom)
                        }
                    }
                }

                Divider()
                    .background(terminalDim)

                // ── Input row ────────────────────────────────────────
                HStack(spacing: 8) {
                    Text(">")
                        .font(.system(.body, design: .monospaced).bold())
                        .foregroundColor(terminalGreen)

                    TextField("", text: $inputText)
                        .font(.system(.body, design: .monospaced))
                        .foregroundColor(terminalGreen)
                        .tint(terminalGreen)
                        .focused($inputFocused)
                        .autocorrectionDisabled()
                        .textInputAutocapitalization(.never)
                        .submitLabel(.send)
                        .onSubmit(submitCommand)
                        .disabled(!engine.isRunning)

                    Button(action: submitCommand) {
                        Image(systemName: "arrow.up.circle.fill")
                            .font(.title2)
                            .foregroundColor(canSubmit ? terminalGreen : terminalDim)
                    }
                    .disabled(!canSubmit)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 10)
                .background(inputBg)
            }
        }
        .onAppear { inputFocused = true }
        .preferredColorScheme(.dark)
    }

    private var canSubmit: Bool {
        engine.isRunning && !inputText.trimmingCharacters(in: .whitespaces).isEmpty
    }

    private func submitCommand() {
        let command = inputText.trimmingCharacters(in: .whitespaces)
        guard !command.isEmpty else { return }

        // Echo the command into the log
        engine.addMessage("> \(command)")
        engine.handleInput(command)
        inputText = ""
    }
}

#Preview {
    ContentView()
}
