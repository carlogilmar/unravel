# Reviewing AI-Generated Code: A Foundation for Unravel

This document develops the theory behind Unravel. It asks three questions:

1. What actually happens in the brain when we read code?
2. Why does AI-generated code create a *new kind* of cognitive problem?
3. What review practices preserve a developer's agency and ownership?

The spine of the argument comes from **Felienne Hermans, _The Programmer's Brain_ (Manning, 2021)**. Other sources are listed at the bottom.

---

## 1. Reading Code Is Harder Than Writing It

In 2000, Joel Spolsky wrote *"Things You Should Never Do, Part I"* — an essay arguing that reading code is fundamentally harder than writing it, and that this asymmetry is why teams keep choosing rewrites over reading what already exists. He framed it as a discipline problem.

Hermans gives the same observation a cognitive explanation. Writing code is a *generative* act: you build a mental model in step with the syntax, so by the time the code exists, the model is already in your head. Reading code is *reconstructive*: the artifact is finished, and you have to rebuild the model that the author had — without their context, without their intermediate states, without their false starts.

AI changes this asymmetry in a brutal way. Before AI, the ratio of code-you-wrote to code-you-read was high enough that the writer's model and the reader's model often overlapped (you were the writer most of the time). AI inverts the ratio. Now the developer is mostly a reader of code that no human ever modeled. There is no original mental model to reconstruct, because no human ever held one.

> **Implication for Unravel.** The product is not a diff viewer. It is a tool for *constructing*, not reconstructing, a mental model of code that was never mentally modeled in the first place.

---

## 2. The Cognitive Architecture (Hermans, Chapters 1–3)

Hermans builds her framework on three memory systems, drawn from cognitive psychology (Atkinson & Shiffrin's model, refined by later work).

### 2.1 Long-Term Memory (LTM)

The permanent store. Holds:
- **Syntactic knowledge**: what `with` does in Elixir, how Python decorators work.
- **Semantic knowledge**: design patterns, idioms, common shapes of solutions.
- **Domain knowledge**: how billing systems work, how OAuth flows go.

LTM is what makes a senior engineer's reading effortless and a junior's reading exhausting. The senior recognizes whole *chunks* (de Groot's chess work, 1965, generalized to programming by Soloway & Ehrlich in the 1980s) where the junior sees raw tokens.

### 2.2 Short-Term Memory (STM)

The buffer. Holds 4–7 items, briefly. When you're reading a function and trying to keep "what does `x` hold right now?" in your head, you're using STM.

### 2.3 Working Memory (WM)

The processor. This is where reasoning happens — where you combine what's in STM with what you pull from LTM to *interpret* the code in front of you.

### 2.4 The Three Cognitive Loads

From John Sweller's **Cognitive Load Theory** (1988):

- **Intrinsic load** — the inherent difficulty of the problem. Cannot be reduced.
- **Extraneous load** — difficulty added by *how* the material is presented. Bad names, inconsistent style, missing context.
- **Germane load** — effort spent building durable mental models that move to LTM.

Good tooling **reduces extraneous load** and **protects germane load**. Bad tooling does the opposite — and worst of all, tooling that "helpfully" summarizes can eliminate germane load entirely, leaving the reader feeling productive while learning nothing.

> **Implication for Unravel.** A summary feature is tempting and probably wrong. Summaries reduce intrinsic load — they trade understanding for speed. The product thesis is to *protect germane load*: keep the work hard enough that something sticks.

---

## 3. Why AI-Generated Code Is Cognitively Unusual

Human-written code carries traces of how it was built: variable names mid-evolution, a comment explaining a weird branch, structure that mirrors the author's path through the problem. These are **beacons** (Brooks 1983, popularized again by Hermans) — small signals that help a reader chunk the code into meaningful units.

AI-generated code is unusually *uniform*. It tends to:

1. **Lack beacons.** Naming is generic-good ("data", "result", "process_item") rather than specifically-good ("unbilled_invoices", "stripe_charge_response"). Every function looks reasonable; few stand out.
2. **Look confident regardless of correctness.** The cadence, the comments, the docstrings — they all signal authority even when the logic is wrong.
3. **Be locally plausible, globally inconsistent.** Each function reads well on its own. Cross-file invariants break silently.
4. **Skip the construction history.** No commit-by-commit evolution to reverse-engineer. The reader sees only the finished artifact, all at once.

The result is a perverse case of **low extraneous load with low germane load**: the code is *easy to read* and *easy to fail to understand*. The reader skims, nods, moves on, and ends up with code in their repository that no human in the building can defend.

This is the agency problem. Not "is the code wrong?" but "if it breaks at 2am, who in this company can fix it?"

---

## 4. A Reading Method for AI Code

Hermans proposes structured techniques for reading human code. They need slight adaptation for AI code.

### 4.1 Two-Pass Reading

- **Pass 1 — Top-down (gist).** Read function signatures, file names, the diff's shape. Form a hypothesis: *"This change appears to add X by modifying Y."* Write the hypothesis down before reading the bodies.
- **Pass 2 — Bottom-up (line-by-line).** Read every changed line. For each, ask: *Is this what I expected? If not, why is the AI doing it this way?* Surprises are the signal.

The order matters. Bottom-up first traps the reader in tokens. Top-down first creates the scaffold that bottom-up fills in.

### 4.2 Roles of Variables (Sajaniemi, 2002)

Jorma Sajaniemi's research at the University of Eastern Finland identified ~11 stereotyped roles that variables play across most code: *fixed value, stepper, most-recent holder, most-wanted holder, gatherer, follower, one-way flag, temporary, organizer, container, walker*. Hermans (Chapter 5) shows that labeling each variable with its role accelerates comprehension.

For AI code, this is doubly useful: it forces the reader to interrogate names that *look* descriptive but might be misleading.

### 4.3 Chunking and Beacon-Hunting

Read the diff looking for beacons: variable names that carry real domain meaning, comments that explain *why* (not what), structural choices that imply intent. Mark beacons. Code that has none is code you have not yet understood — even if you "read" it.

### 4.4 The Three Ownership Tests

Borrowed from teaching pedagogy and adapted here:

1. **The modification test.** *Could I change this safely?* If you can't predict the blast radius of a one-line change, you don't own the code.
2. **The deletion test.** *Could I argue for removing this?* If the answer is "I don't know what would break," you're a passenger.
3. **The explanation test.** *Could I explain this to a teammate in plain language, without rereading?* This is the strongest signal that comprehension has moved from WM to LTM.

A review session is complete not when every line is read, but when the developer can pass at least the modification test for every changed hunk.

### 4.5 Read Tests Separately from Implementation

AI often writes tests that mirror the implementation rather than the *specification*. Reading them together creates an illusion of correctness. Read the test alone, predict the implementation, then check.

---

## 5. Agency and Ownership

Agency in software is the felt ability to act on the code. Ownership is the durable knowledge that backs that ability. The two collapse together when AI generates code that the developer cannot defend.

Hermans frames programming expertise as **moving knowledge into LTM**. Agency requires that some non-trivial fraction of the codebase live in your LTM, not just on disk. AI accelerates code-on-disk growth without contributing to LTM growth. Without a deliberate practice to close that gap, the codebase grows faster than the team's collective LTM, and the team's effective ownership shrinks even as their output grows.

This is the real cost of "vibe coding" — not bugs (those are findable), but the slow erosion of the team's capacity to reason about their own system.

The defense is not "AI bad." The defense is a review discipline that converts read-code into known-code at roughly the rate AI produces it.

---

## 6. Implications for Unravel's Design

Translating the theory into product decisions:

| Principle | Feature implication |
|---|---|
| Protect germane load | No AI summaries in v1. The reader does the work. |
| Top-down then bottom-up | Two view modes: file list (gist) then per-hunk reading mode (detail). |
| Mark understood | The "✓ understood" checkbox is the core interaction, not a side feature. |
| Beacons matter | Allow per-line notes; treat them as first-class. Notes are externalized beacons. |
| Roles of variables (v2) | Future: let the user tag variables with Sajaniemi roles to make naming critique structural. |
| Modification test | Future: a "predict the effect of changing this line" prompt before allowing the mark. |
| Construction history is missing | Future: per-hunk "why" prompts that the developer answers, not the AI. |
| Ownership decays with time | Track a "last reviewed at" timestamp per file. Surface drift over weeks. |

The negative space matters too. Things Unravel should **not** do in v1:

- **Auto-explain code.** Defeats germane load.
- **Auto-mark hunks as trivial.** The decision of what is trivial is itself the review.
- **Gamify with streaks or scores.** Cheapens the act. Marking understood should feel like signing your name.

---

## 7. References

### Primary

- **Felienne Hermans.** *The Programmer's Brain: What every programmer needs to know about cognition.* Manning, 2021.
- **Felienne Hermans' blog and Code Reading Clubs.** [felienne.com](https://www.felienne.com), [codereading.club](https://www.codereading.club) — a community practice of reading code together, directly relevant to the kind of discipline Unravel wants to enable.

### Cognitive science foundations

- **John Sweller.** *Cognitive Load During Problem Solving: Effects on Learning.* Cognitive Science, 1988. Origin of Cognitive Load Theory.
- **Adriaan de Groot.** *Thought and Choice in Chess.* 1965. The chess-expertise studies that inspired most later work on expert chunking.
- **Jorma Sajaniemi.** *An Empirical Analysis of Roles of Variables in Novice-level Procedural Programs.* IEEE Symposia on Human-Centric Computing, 2002. Origin of the roles-of-variables framework.
- **Elliot Soloway & Kate Ehrlich.** *Empirical Studies of Programming Knowledge.* IEEE TSE, 1984. Foundational work on programming plans and beacons.
- **Ruven Brooks.** *Towards a theory of the comprehension of computer programs.* International Journal of Man-Machine Studies, 1983. Early treatment of beacons in code reading.

### Software engineering practice

- **Joel Spolsky.** *"Things You Should Never Do, Part I."* Joel on Software, 2000. The classic case for reading over rewriting. [joelonsoftware.com](https://www.joelonsoftware.com/2000/04/06/things-you-should-never-do-part-i/)
- **Michael Feathers.** *Working Effectively with Legacy Code.* Prentice Hall, 2004. Practical techniques for reading and modifying code you didn't write — directly applicable to AI-generated code, which is "legacy on arrival."
- **John Ousterhout.** *A Philosophy of Software Design.* Yaknyam Press, 2018. On complexity, naming, and the cost of poor abstraction — useful vocabulary for critiquing AI output.
- **Steve McConnell.** *Code Complete*, 2nd ed. Microsoft Press, 2004. Chapters on reading and self-review remain useful.
- **Martin Fowler.** *Refactoring*, 2nd ed. Addison-Wesley, 2018. The catalog is a vocabulary of safe edits — a basis for the "modification test."

### AI-specific reading (caveat: fast-moving area)

- **Simon Willison's blog.** [simonwillison.net](https://simonwillison.net) — ongoing, careful writing on what LLMs do and don't do when generating code. A good antidote to hype.
- **Addy Osmani.** Public writing on AI-assisted development workflows.

---

## 8. One-Line Summary

> Unravel exists because AI produces code faster than humans can move it into long-term memory, and the only known fix is a deliberate reading practice that the tool refuses to short-circuit.
