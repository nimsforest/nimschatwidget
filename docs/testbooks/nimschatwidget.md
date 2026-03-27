# NimsChatWidget Testbook

Validates the `nimschatwidget` package -- thin iframe launcher that opens the webchat embed mode.

## Prerequisites

- nimschatwidget running (standalone or embedded)
- Webchat available at the configured `webchatURL` (e.g., `https://webchat.nimsforest.mynimsforest.com`)
- A host page that loads the widget script

## 1. Widget JS endpoint

```bash
curl -s https://chatwidget.nimsforest.mynimsforest.com/widget | head -c 30
# Expected: (function() {
```

Verify response headers:

```bash
curl -sI https://chatwidget.nimsforest.mynimsforest.com/widget
# Expected: Content-Type: application/javascript; charset=utf-8
# Expected: Access-Control-Allow-Origin: *
# Expected: Cache-Control: public, max-age=300
```

## 2. Health endpoint

```bash
curl -s https://chatwidget.nimsforest.mynimsforest.com/health
# Expected: {"status":"ok"}
```

## 3. CORS preflight

```bash
curl -sI -X OPTIONS https://chatwidget.nimsforest.mynimsforest.com/widget
# Expected: 204 No Content
# Expected: Access-Control-Allow-Origin: *
```

## 4. Host page integration

Embed the widget on any page:

```html
<script>
  window.nimschatwidgetConfig = {
    webchatURL: 'https://webchat.nimsforest.mynimsforest.com',
    context: 'Page: Homepage\nUser: test'
  };
</script>
<script src="https://chatwidget.nimsforest.mynimsforest.com/widget"></script>
```

## 5. Browser validation -- button

1. Open the host page
2. **Verify**: green floating chat button (56px circle, #4AA847) in bottom-right corner
3. Hover over button
4. **Verify**: button scales up slightly, background darkens to #3d8f3c

## 6. Browser validation -- iframe panel

1. Click the green button
2. **Verify**: iframe panel appears above the button (bottom: 96px, right: 24px)
3. **Verify**: panel is 400x600px with 16px border-radius and drop shadow
4. **Verify**: iframe loads the webchat embed URL with session and context query parameters
5. **Verify**: close button (x) appears in top-right of the panel
6. Click close button
7. **Verify**: panel hides, button remains visible
8. Click button again
9. **Verify**: same iframe reopens (no reload, sends postMessage with updated context)

## 7. Session persistence

1. Open widget, note the iframe src URL -- extract the `session=` parameter value
2. Close browser tab, reopen the page
3. Open widget again
4. **Verify**: same session ID in the iframe URL (stored in localStorage as `ncw-session`)

## 8. Context updates

1. Open widget (iframe loads)
2. In browser console, update context:
   ```js
   window.nimschatwidgetConfig.context = 'Updated context: new page';
   ```
3. Close and reopen the widget panel
4. **Verify**: postMessage with `{type: 'context', context: 'Updated context: new page'}` is sent to the iframe (check via browser DevTools > Console, or add a message listener in the webchat)

## 9. Mobile layout

1. Open on mobile or resize browser below 640px width
2. **Verify**: chat button is 48px, positioned 16px from edges
3. Click button
4. **Verify**: iframe panel goes full-width and full-height (covers entire viewport)
5. **Verify**: close button still accessible

## 10. Double-init prevention

1. Load the widget script twice on the same page
2. **Verify**: only one button and one panel appear (the `ncw-root` ID check prevents duplicates)

## 11. Missing webchatURL

1. Load widget without setting `webchatURL`:
   ```html
   <script>window.nimschatwidgetConfig = {};</script>
   <script src="https://chatwidget.nimsforest.mynimsforest.com/widget"></script>
   ```
2. Click the button
3. **Verify**: nothing happens, console shows `[nimschatwidget] webchatURL not configured`

## Pipeline trace

The widget itself has no backend pipeline. The full message flow is:

```
User clicks button
  -> iframe opens webchat /embed?session={id}&context={ctx}
  -> webchat handles all chat UI, messaging, SSE, nim routing
  -> widget only manages the iframe visibility and context updates via postMessage
```
