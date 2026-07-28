// A deliberately small markdown subset for note bodies — headings, bold, italic,
// inline code, bullet and numbered lists, blockquotes. No dependency, and no raw
// HTML passthrough: the source is escaped first, so a note that happens to contain
// "<script>" renders as that text rather than executing.

const ESCAPES = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }

function escapeHtml(text) {
  return text.replace(/[&<>"']/g, (c) => ESCAPES[c])
}

// Applied to already-escaped text, so the patterns only ever see literal markers.
// The emphasis patterns require a non-space right after the opening marker and
// demand a word boundary before it, so prose like "3 * 4" and identifiers like
// snake_case_word survive untouched.
function inline(text) {
  return text
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^\s*][^*]*?)\*\*/g, '<strong>$1</strong>')
    .replace(/(^|[^*\w])\*([^\s*][^*]*?)\*/g, '$1<em>$2</em>')
    .replace(/(^|[^_\w])_([^\s_][^_]*?)_/g, '$1<em>$2</em>')
}

function listItem(line) {
  const bullet = line.match(/^\s*[-*+]\s+(.*)$/)
  if (bullet) return { ordered: false, text: bullet[1] }
  const numbered = line.match(/^\s*\d+[.)]\s+(.*)$/)
  if (numbered) return { ordered: true, text: numbered[1] }
  return null
}

// fts5's snippet() wraps matches in <mark> but leaves the surrounding text exactly
// as it was indexed, and note bodies are indexed raw. Escape the lot, then let the
// highlight markers back through — nothing else survives as markup.
export function renderExcerpt(text) {
  return escapeHtml(text ?? '')
    .replace(/&lt;mark&gt;/g, '<mark>')
    .replace(/&lt;\/mark&gt;/g, '</mark>')
}

export function renderMarkdown(source) {
  if (!source?.trim()) return ''

  const out = []
  let list = null

  const closeList = () => {
    if (list) {
      out.push(list.ordered ? '</ol>' : '</ul>')
      list = null
    }
  }

  for (const raw of escapeHtml(source).split('\n')) {
    const line = raw.trimEnd()

    if (!line.trim()) {
      closeList()
      continue
    }

    const item = listItem(line)
    if (item) {
      if (!list || list.ordered !== item.ordered) {
        closeList()
        list = { ordered: item.ordered }
        out.push(item.ordered ? '<ol>' : '<ul>')
      }
      out.push(`<li>${inline(item.text)}</li>`)
      continue
    }

    closeList()

    const heading = line.match(/^(#{1,4})\s+(.*)$/)
    if (heading) {
      const level = Math.min(heading[1].length + 2, 6)
      out.push(`<h${level}>${inline(heading[2])}</h${level}>`)
      continue
    }

    const quote = line.match(/^&gt;\s?(.*)$/)
    if (quote) {
      out.push(`<blockquote>${inline(quote[1])}</blockquote>`)
      continue
    }

    out.push(`<p>${inline(line)}</p>`)
  }

  closeList()
  return out.join('')
}
