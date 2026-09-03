# Playwright video recording of an interactive flow + Next.js SPA exploration

Combines the two techniques that made the cypherspy.com "Verify CypherCoin / enter token"
task work end-to-end and deliverable as a WA video (Aug 2026).

## When to use
- User asks to "record the flow and send me a video" of any interactive browser interaction
  (click a button → popup → fill a field → submit → observe result).
- Or: the site is a Next.js / React SPA where `curl`/static fetch yields 308 + tiny/empty
  body (content rendered client-side, hidden behind navigation menus or popups that only
  appear after clicking).

## 1. Critical SPA pitfall: `networkidle` rarely settles — use `domcontentloaded` + wait
Next.js apps keep a polling/websocket or prefetch connection alive, so
`wait_until='networkidle'` blocks until timeout even though the page is fully usable.
```python
await page.goto(url, wait_until='domcontentloaded', timeout=45000)
await page.wait_for_timeout(4000)   # give client-side render/hydration a moment
```
A 308 from `curl` on a path like `/professionals` is just Next.js trailing-slash redirect —
follow it in the browser; it's not a block.

## 2. Explore the SPA's DOM before recording (find the hidden button + popup)
Buttons and popups may not be visible until you click something. Dump everything, then click
through, dumping each time:
```python
links = await page.eval_on_selector_all('a', 'els => els.map(e => ({t: e.innerText.trim().slice(0,40), h: e.href}))')
btns  = await page.eval_on_selector_all('button', 'els => els.map(e => e.innerText.trim().slice(0,60))')
# find any text containing verify/token anywhere in the DOM:
hits = await page.evaluate('''() => {
  const w = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT); const r=[]; let n;
  while (n = w.nextNode()) if (/verify|token/i.test(n.textContent)) r.push(n.parentElement.tagName+': '+n.textContent.trim().slice(0,100));
  return r; }''')
```
The "verify button" may be named differently than expected (e.g. the user said "verify token"
but the element is labeled **"Verify CypherCoin"** and appears only after you open a tool).
Click `Open Tool` / the action link, then re-dump to reveal the popup with the submit button.
List the visible buttons after opening the modal to find a bare "Verify" next to "Cancel".

## 3. Record the whole flow as video (Playwright webm)
```python
ctx = await browser.new_context(
    user_agent='...Chrome/126...', viewport={'width':1440,'height':900}, locale='en-US',
    record_video_dir='/root/cypher_video', record_video_size={'width':1440,'height':900},
)
page = await ctx.new_page()
# ... drive the flow: goto -> click -> fill -> click ...
await page.screenshot(path='/root/flow_final.png', full_page=False)
await ctx.close()                       # MUST close context before reading the video
vpath = await page.video.path()         # /root/cypher_record/page@<id>.webm
shutil.copy(vpath, '/root/flow.webm')   # webm is WhatsApp-acceptable
```
- Recording auto-starts on the first navigation; it produces a `.webm`.
- **Read `page.video.path()` only AFTER `await ctx.close()`** — before close the file isn't
  finalized. Copy it to a stable path so you have a predictable handle.
- The whole step sequence lives inside the same script (goto→click→fill→click→wait), with
  prints between steps so you can see where it was when it ended.

## 4. Capturing the outcome honestly
The recording captures exactly what happened — including a *failure*. In this case the token
`123456789098` was rejected by the site's own validation: the modal came back red with
"Security verification failed. Please try again." Record and REPORT the real result; do not
retouch the video or fabricate a success. Send the webm via the Baileys `wa_send_file`
(`path=` param, caption + mime video/webm). The user gets an honest 1.9MB clip.

## Pitfalls
- `record_video_size` must match a real viewport to avoid letterboxing.
- Don't guesionlabels — dump the DOM to find the actual button text before recording.
- `networkidle` on an SPA = hang; always `domcontentloaded` + explicit wait_for_timeout.
- Sending a `.webm` video works in WA but is large — keep the recording short.