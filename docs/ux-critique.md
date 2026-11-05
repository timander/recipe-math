# UX expert critique of v2

Written 2026-11-04 by a separate Claude agent playing a blunt UX/UI and educational-game designer. It read the code, studied screenshots, and played stations 3 and 5 at Chromebook (1366×768) and iPhone sizes.

## The roast
1. **The home screen puts the final boss on top.** The list is reversed, so on a phone the first card is Chef de Cuisine. Dish Pit is four thumb-scrolls down, and a 15-year-old taps the first big button they see.
2. **A wall of rules comes before any play.** Nobody reads four bullets about half-plates. They learn the rules by losing plates.
3. **One mis-tap on a choice costs a whole plate.** Tapping a choice serves it immediately, and on a Chromebook any stray a–e or 1–5 keypress serves too. Fix the Mistake's options F–I can't be reached from the keyboard.
4. **Fix the Mistake is a matching game spread across two cards.** The original is a small side card; the choices are a list in another column, in a different format. To check one line, a student has to find its match, scale it, *and* decode "3 c + ¼ c + 2 T" back into 3⅜ c. That's three skills in one click, and the rows don't line up. On a phone the original scrolls away.
5. **Some "mistakes" don't need the comparison at all.** The wrong-tool variant can be spotted without the original. The milk option leaked "(liquid cup)" twice.
6. **The glass shows up for ¾ t of vanilla,** pushing the spoons about 400px below the bowl on a phone.
7. **The glass's fill lines are hard to hit:** 2px lines, roughly 16px targets, 0.7rem labels.
8. **On phones, the sticky HUD wraps to two rows** and covers the rail and the card.
9. **A correct answer strikes through the options you never picked,** so it looks like guessing and the text is hard to read during review.
10. **Tips are a number with no meaning.** The 📋 rail emoji is decoration without a legend.
11. **The bowl is a pile of chips:** "1 T ×", "1 T ×", "1 T ×", with no grouping or counts.

**What's good:** the Socratic feedback ("better than most commercial edtech"), the knockout "Study these" review, paper parity with seeded keys, and the real-tool drawer with "any exact combo counts."

## Top fixes (impact ÷ effort)
1. Dish Pit first; new students see one big "start here" button. (S)
2. Choices select, then confirm; keys only select, Enter serves. (S)
3. Don't strike through untouched options on a correct answer. (S)
4. Rebuild Fix the Mistake as one ledger table inside the ticket: ingredient | original | ×f cook wrote | tap the row. Show scaled amounts as a single quantity, not scoop combos. Drop the wrong-tool variant. (M)
5. Turn the bowl into a sticky tray on phones; group repeats ("1 T ×3") with −/+; make the glass a tile that expands; make fill lines 28px targets. (M)
6. Make the HUD one row on phones (✕ instead of "Leave"). (S)
7. Add a beginner mode: single conversions, no clock, extra plates. (M)
8. Weight the factors: ×½ and ×2 about 70%, ×3 and ×4 about 25%. (S)
9. Teach the rules in context instead of a wall of text. (S)
10. Give tips a purpose (cosmetic unlocks or personal bests), or drop them. (M)

## Game feel
Streaks ("🔥 Hot line ×3"); juice on correct (ticket slides off, chip pops) and on wrong (plate falls from the HUD); a "MISE EN PLACE ✓" stamp on finished cards; auto-advance clean first-try answers after about 1.2s; mastery rings; a Rush mode unlocked at 3★. Avoid public leaderboards, loud default sounds, and a red clock on a first visit; default the timer off for beginners.

## Accessibility and classroom checklist
- Disabled buttons at 45% opacity fail contrast; glass labels need at least 0.85rem.
- Touch targets of at least 44px (fill lines, bowl chips, small buttons).
- Move focus to the feedback after a wrong answer.
- Announce dropped plates in text; add `aria-pressed` on the active glass line; give each bowl group a label with its count.
- Cover all motion under reduced-motion.
- Print Fix the Mistake as the same ledger table.
- Classroom links: `?rank=1`, and show the version number on screen.
