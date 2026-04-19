/**
 * Replaces [[wikilinks]] in text with HTML spans.
 * [[Page Name]] → <span class="wikilink">Page Name</span>
 * [[Page Name|Display Text]] → <span class="wikilink">Display Text</span>
 */
export function renderWikilinks(text: string): string {
	return text.replace(/\[\[([^\]]+)\]\]/g, (_, content: string) => {
		const parts = content.split('|');
		const display = parts.length > 1 ? parts[1].trim() : parts[0].trim();
		return `<span class="wikilink">${escapeHtml(display)}</span>`;
	});
}

function escapeHtml(s: string): string {
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
