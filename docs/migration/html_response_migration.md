# Migration Guide: HTML Responses from AI API

## Overview
The Bible AI API has been updated to return **semantic HTML** in the response `text` field for AI queries, instead of plain text or variable Markdown. Additionally, the API now preserves HTML structure in the context verses sent to the LLM, enabling better handling of poetry and formatting.

## Changes
- **Endpoint:** `/query` (when `prompt` is provided)
- **Response Field:** `text` (inside the JSON response)
- **Old Format:** Plain text or Markdown (inconsistent).
- **New Format:** Semantic HTML string (e.g., `<p>God loves you.</p>`).

## Client Migration Instructions

### For Web Applications (e.g., Discipleship Journal)
This change should be largely beneficial. You can render the returned HTML directly using your framework's safe HTML rendering method (e.g., `dangerouslySetInnerHTML` in React, `v-html` in Vue).
- **Action:** Ensure your frontend renders the `text` field as HTML.

### For Telegram Bots (e.g., ScriptureBot)
Telegram supports a strictly limited subset of HTML tags. The new API may return tags that Telegram does not support (e.g., `<div>`, `<p>`, `<h1>`, `<ul>`). Sending these raw tags to Telegram will result in them being displayed as text or causing parsing errors if you use `parse_mode="HTML"`.

**Supported Telegram Tags:** `<b>`, `<strong>`, `<i>`, `<em>`, `<u>`, `<ins>`, `<s>`, `<strike>`, `<del>`, `<span class="tg-spoiler">`, `<a>`, `<code>`, `<pre>`.

**Action:** You must preprocess the HTML string before sending it to Telegram.

#### Recommended Parsing Logic (Pseudocode):

```python
def convert_html_for_telegram(html_content):
    # 1. Convert block-level elements to newlines
    html_content = html_content.replace("<p>", "").replace("</p>", "\n\n")
    html_content = html_content.replace("<br>", "\n")

    # 2. Convert headings to Bold
    import re
    html_content = re.sub(r'<h[1-6]>(.*?)</h[1-6]>', r'\n\n<b>\1</b>\n', html_content)

    # 3. Convert lists to bullet points
    html_content = html_content.replace("<ul>", "").replace("</ul>", "")
    html_content = html_content.replace("<li>", "• ").replace("</li>", "\n")

    # 4. Strip unsupported tags (keeping content) like div, span (unless generic)
    # You might use a library like BeautifulSoup to strip specific tags while keeping children.

    # 5. Sanitize remaining tags to ensure only Telegram-supported tags remain.
    # Be careful of unclosed tags which Telegram rejects.

    return html_content.strip()
```

### Context for LLM
The API no longer strips HTML from the Bible verses sent to the LLM. This allows the LLM to "see" the structure of the text (e.g., poetry lines, headings) and provide better, context-aware answers. This is an internal improvement and shouldn't require client-side changes, but you may notice the AI referencing formatting or structure more accurately.
