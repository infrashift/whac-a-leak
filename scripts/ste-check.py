#!/usr/bin/env python3
"""Check documentation against ASD-STE100 (Simplified Technical English) basics.

Errors (exit 1): banned words and phrases, sentences over 25 words.
Warnings (exit 0): likely passive voice.
Run from the repository root: python3 -I scripts/ste-check.py
"""
import pathlib
import re
import sys

MAX_WORDS = 25
BANNED = [
    "should", "will", "won't", "would", "please", "simply", "just", "very",
    "really", "easily", "is expected to", "are expected to", "can be seen",
]
BANNED_RE = re.compile(r"\b(" + "|".join(re.escape(b) for b in BANNED) + r")\b", re.I)
PASSIVE_RE = re.compile(r"\b(is|are|was|were|be|been|being)\s+\w+ed\b", re.I)
SENTENCE_END = re.compile(r"(?<=[.!?])\s+(?=\S)")


def files(root):
    yield from (root / n for n in ("README.md", "PLAN.md", "CLAUDE.md") if (root / n).exists())
    yield from sorted((root / "docs/src/content/docs").rglob("*.md*"))
    for p in sorted((root / ".okfw").rglob("*.md")):
        rel = p.relative_to(root / ".okfw")
        if rel.parts[0] != "daily" and p.name != "index.md":
            yield p


def clean(line):
    """Drop inline code, link targets, HTML tags and emphasis markers."""
    line = re.sub(r"`[^`]*`", "CODE", line)
    line = re.sub(r"\]\([^)]*\)", "]", line)
    line = re.sub(r"<[^>]+>", " ", line)
    line = re.sub(r"https?://\S+", "URL", line)
    return line.replace("**", "").replace("*", "")


def prose_lines(text):
    """Yield (line number, text) for prose lines only."""
    lines = text.split("\n")
    i, n = 0, len(lines)
    if lines and lines[0].strip() == "---":
        i = 1
        while i < n and lines[i].strip() != "---":
            i += 1
        i += 1
    fence = False
    while i < n:
        raw = lines[i]
        s = raw.strip()
        i += 1
        if s.startswith("```"):
            fence = not fence
            continue
        if fence or not s:
            continue
        if s.startswith(("#", "|", "<", "import ", "export ", "/>", "{")):
            continue
        if s == "<!-- BEGIN okfw -->":
            return
        yield i, s


def words(sentence):
    return [w for w in re.split(r"\s+", sentence) if re.search(r"[A-Za-z0-9]", w)]


def check(path, text, report):
    if path.name == "CLAUDE.md":
        text = text.split("<!-- BEGIN okfw -->")[0]
    # Join wrapped lines of one paragraph or list item so sentences are whole.
    para, start = [], 0

    def flush():
        if not para:
            return
        joined = clean(" ".join(para))
        joined = re.sub(r"^\s*([-*+]|\d+\.)\s+(\[[ x]\]\s+)?", "", joined)
        for sent in SENTENCE_END.split(joined):
            if len(words(sent)) > MAX_WORDS:
                report("error", path, start, f"sentence has {len(words(sent))} words (max {MAX_WORDS}): {sent[:60]}…")
        for m in BANNED_RE.finditer(joined):
            report("error", path, start, f"banned word or phrase: {m.group(0)!r}")
        for m in PASSIVE_RE.finditer(joined):
            report("warning", path, start, f"possible passive voice: {m.group(0)!r}")

    prev = None
    for num, s in prose_lines(text):
        new_item = re.match(r"^([-*+]|\d+\.)\s", s) is not None
        if prev is None or num != prev + 1 or new_item or s.startswith(">"):
            flush()
            para, start = [], num
        para.append(s.lstrip("> "))
        prev = num
    flush()


def main():
    root = pathlib.Path.cwd()
    counts = {"error": 0, "warning": 0}

    def report(level, path, line, msg):
        counts[level] += 1
        print(f"{path.relative_to(root)}:{line}: {level}: {msg}")

    for p in files(root):
        check(p, p.read_text(encoding="utf-8"), report)
    print(f"ste-check: {counts['error']} error(s), {counts['warning']} warning(s)")
    return 1 if counts["error"] else 0


if __name__ == "__main__":
    sys.exit(main())
