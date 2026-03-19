# NimsChatWidget Testbook

Validates the `nimschatwidget` package — embeddable nim chat for any NimsForest surface.

## Prerequisites

- A NimsForest app integrating nimschatwidget (e.g., nimsforestissue)
- Forest pipeline running (webhook -> tree -> treehouse -> nim -> agentclaudecode)
- Wind (NATS) connected
- `webhook_url` configured in the embedding app's config

## 1. Package integration

```go
// Minimal integration in any NimsForest app:
source := nimschatwidget.NewSource(webhookURL, "myapp")
songbird := nimschatwidget.NewSongbird(wind)
songbird.Start()
mux.Handle("/admin/chat/", http.StripPrefix("/admin/chat", nimschatwidget.Handler(source, songbird)))
```

Verify: app starts without errors, logs show `Catching leaves on subject: song.nimschatwidget.>`

## 2. Widget endpoint

```bash
# Serves JS content
curl -s https://<app>/admin/chat/widget | head -c 30
# Expected: (function() {
```

## 3. Nims list

```bash
curl -s https://<app>/admin/chat/nims | python3 -m json.tool | head -5
# Expected: JSON array of {name, role} objects
```

## 4. Send message

```bash
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"session_id":"test-1","target_nim":"nimble","text":"hello","context":"Test context"}' \
  https://<app>/admin/chat/send
# Expected: {"status":"ok"}
```

## 5. SSE events

```bash
# In terminal 1: listen for events
curl -s -N "https://<app>/admin/chat/events?session=test-1"

# In terminal 2: send a message
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"session_id":"test-1","target_nim":"nimble","text":"say hello","context":""}' \
  https://<app>/admin/chat/send

# Expected in terminal 1: data: {"text":"...","source":"nimble"}
```

## 6. Context passthrough

```bash
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"session_id":"ctx-test","target_nim":"nimble","text":"what context do you see?","context":"Issue #42: Fix login bug\nStatus: open\nPriority: high"}' \
  https://<app>/admin/chat/send

# Listen for response:
timeout 60 curl -s -N "https://<app>/admin/chat/events?session=ctx-test"
# Expected: nim response should reference Issue #42 / login bug
```

## 7. Browser validation

1. Open the app's admin page where the widget is embedded
2. **Verify**: green floating chat button in bottom-right corner
3. Click the button
4. **Verify**: slide panel opens from right with nim selector dropdown
5. **Verify**: dropdown lists nims (neo, nimble, nurture, etc.)
6. Select a nim, type a message, press Enter
7. **Verify**: user message appears as light green bubble (#DCF5DB) on right
8. **Verify**: bouncing dot typing indicator appears (3 green dots)
9. **Verify**: nim response appears as white bubble with border on left, with nim name label
10. Click X to close panel
11. **Verify**: panel slides closed

## 8. Visual design alignment

The widget should match nimsforestwebchat's design:

| Element | Expected |
|---------|----------|
| User bubble | #DCF5DB background, #1E3A1C text, rounded with small bottom-right corner |
| Nim bubble | White background, #6B5B4E text, 1px #EDE9E5 border, small bottom-left corner |
| Body font | Georgia, "Source Serif 4", serif |
| UI font | Inter, DM Sans, system sans-serif |
| Send button | 40px circle, #4AA847 green, white arrow SVG |
| Input | #F0F3ED background, #E2DDD8 border, 12px radius |
| Typing indicator | 3 bouncing dots, #A8D5A2 color, 1.4s animation |
| Message area | #F8FAF5 background |
| Header | White background, #EDE9E5 bottom border |

## 9. SSE reconnect

1. Open chat widget in browser
2. Restart the app container
3. **Verify**: SSE reconnects within ~3 seconds (check Network tab)
4. Send a new message
5. **Verify**: message sends and response arrives

## 10. Mobile layout

1. Open on mobile or resize browser < 768px
2. **Verify**: chat button is 48px, positioned 16px from edges
3. **Verify**: panel opens as full-width overlay

## 11. Songbird cleanup (chirp delivery)

1. Open chat widget, note the session ID
2. Close the browser tab
3. Check server logs — Songbird should log the listener removal (no orphaned chirp channels)
4. No memory leak from abandoned SSE connections

## Pipeline trace

Full message flow to verify end-to-end:

```
Widget POST /send
  -> Source POST to forest webhook (http://46.225.164.179:8081/webhooks/chatwidget)
  -> River (river.chat.widget)
  -> Tree (message-chat) — parses, sets reply_subject=song.nimschatwidget.{session}
  -> message.incoming
  -> TreeHouse (message_router.lua) — routes to message.{target_nim}
  -> Nim (renders prompt with {{.context}})
  -> AgentBrain -> agent.work.ai.nimble.chat -> agentclaudecode
  -> Result -> Nim publishes to song.nimschatwidget.{session}
  -> Songbird catches -> chirps via SSE -> Widget displays
```
