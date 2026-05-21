# Review code in AI era

> Carlo's intro

Nowadays we are increasing the speed of development thanks to LLM's and code generation, this is another level of building software and automate our workflows. It feels more like to be a copilot rather than being a software writter.

Before I used my terminal and vim to write software, now I need a subscription to a LLM, send a prompt and let them finish my task, to review the code and then integrate, it's a very different approach with many challenges.

---

## Where the industry stands in 2026

Between Andrej Karpathy's February 2025 "vibe coding" tweet and today, the software industry has run a year-long, mostly unregulated experiment in letting LLMs author production code. The headline numbers from the 2025 DORA report make the scale clear: **90% of surveyed engineers use AI at work, and over 80% feel more productive**. But the same report, and the empirical studies that followed it, also describe a less comfortable reality:

- AI adoption is now correlated with **higher software delivery throughput** but **lower software delivery stability**, unless the team has mature testing, version control, and feedback loops in place. (2025 DORA)
- In a randomized controlled trial of 16 experienced open-source developers across 246 real tasks, **AI tools made them 19% slower** while making them feel 24% faster. The gap between perception and reality is the most striking single finding of the year. (METR, July 2025)
- A five-year analysis of 211M lines of code found **code clones with 5+ duplicated lines grew 8× in 2024**, refactoring activity dropped from 25% of changed lines (2021) to under 10% (2024), and the share of code revised within two weeks of commit grew from 3.1% to 5.7%. (GitClear, 2025)
- **45% of AI-generated code samples introduce OWASP Top-10 vulnerabilities**, and AI code carries a 2.74× higher vulnerability rate than human-written code. (Veracode, 2025)
- Roughly **20% of packages an LLM recommends do not exist** (~22% for open models, ~5% for commercial) — creating the new attack surface known as "slopsquatting." (USENIX Security, 2025)
- An Anthropic randomized trial measured **a 17-point drop in skill mastery** when developers used AI assistance during learning tasks (50% quiz scores with AI vs. 67% without).

Read together, these results say something specific: AI is amplifying both the strengths and the weaknesses of the engineering system it lands in. Where the surrounding practices are strong, AI speeds delivery. Where they are weak, AI accelerates the production of plausible-but-fragile code, and quietly eats away at the team's grasp of its own system. The 2026 conversation is no longer "does AI help?" but "what *practice* makes AI a net positive on a real codebase?"

---

## Four frames that emerged in 2025

The community converged on roughly four mental models for working with AI coding tools. They are not mutually exclusive — most senior engineers fluidly switch between them — but their failure modes differ.

### 1. Vibe coding (Karpathy, Feb 2025)

> "There's a new kind of coding I call *vibe coding*, where you fully give in to the vibes, embrace exponentials, and forget that the code even exists."
> — Andrej Karpathy

The original framing: speak to the agent, accept whatever it produces, never inspect the code. Karpathy intended it for *throwaway* and *exploratory* work. The term escaped containment within weeks and became shorthand for any prompt-driven development — including production work, where its assumptions don't hold.

**Best for:** prototypes, weekend hacks, learning a new framework, throwaway scripts.
**Failure mode:** unreviewed code reaching production. Both Karpathy's own follow-ups and Simon Willison's "vibe coding rocks but is risky" essay are explicit that this is *not* a production-engineering practice.

### 2. Vibe engineering (Willison, Oct 2025)

Simon Willison coined this term to describe what experienced engineers actually do when they use LLMs seriously: combine the agent's speed with the engineer's discipline. The slogan is **"if you reviewed, tested, and understood every line, that's not vibe coding — that's using an LLM as a typing assistant."**

**Best for:** day-to-day professional work where the engineer remains accountable for the diff.
**Practice stack:** specifications up front, automated tests, intentional review, documentation, code-reading discipline.

### 3. Augmented coding (Beck, Sep 2025)

Kent Beck — creator of XP and TDD, with five decades of programming behind him — framed his own daily use of agents as **augmented coding**: AI amplifies the developer rather than replacing them. The cleanest practical demonstration is his BPlusTree3 library (Rust + Python), built end-to-end with an agent but anchored in TDD discipline. Beck's distinction is essentially: *you decide what is correct; the AI helps you get there faster*.

### 4. Spec-driven development (SDD; broadly adopted 2026)

The most consequential 2026 shift. SDD treats an **executable, version-controlled specification as the single source of truth**, with code as derived artifact. The workflow is typically four phases — Specify → Plan → Tasks → Verify — with a *separate* verifier agent checking the implementer's work rather than letting the same agent grade its own homework. Early adopters at GitHub and AWS report **3–10× higher first-pass success rates** on non-trivial tasks. By mid-2026, GitHub Spec Kit, AWS Kiro, Tessl, OpenSpec, BMAD, and Google Antigravity all ship their own SDD flavor.

Addy Osmani's "Beyond Vibe Coding" essay and the O'Reilly piece "How to Write a Good Spec for AI Agents" argue that **the AI is not the bottleneck — the specification is**. GitHub's own analysis of 2,500+ agent configuration files backs this.

---

## What we have empirically learned

Three findings keep recurring across independent studies. They matter because they are surprising — they are not what most teams *think* is happening.

### 3.1 Developers systematically over-estimate AI speedup

The METR result (19% slower while feeling 24% faster) is the most-cited number of the year. The same pattern appears in Anthropic's skill-mastery study (lower scores, higher confidence) and in Microsoft/Carnegie Mellon's cognitive-offloading work. The shared root cause is that AI removes the *friction* of writing code without removing the *work* of understanding it — and friction is what we use as a proxy for effort. Without friction, developers feel productive while the actual time-to-merged-PR grows.

> **Implication.** "Feels faster" is not evidence of faster. The only honest signal is shipped, working, *understood* code. This is the empirical case for tools that protect comprehension as a first-class goal.

### 3.2 Quality decays at a measurable rate when AI volume outpaces review discipline

GitClear's longitudinal data is the cleanest evidence: as AI authorship rose from 2021 to 2024, the *shape* of the codebase changed in characteristic ways — more clones, less refactoring, more two-week churn. None of these are bugs in the AI; they are bugs in the *practice* around the AI. The AI optimizes for "plausible new code." Humans, historically, had also been doing the work of "reusing and refactoring existing code." When the second behavior is no longer present, the codebase still grows, just less coherently.

> **Implication.** "AI for new lines" without "humans for refactoring and reuse" is a recipe for code-clone debt. The 2024 inflection point — where new repeated code outpaces refactoring for the first time in the GitClear dataset — is the canary.

### 3.3 The "70% problem" and the cost of the last 30%

Addy Osmani's working observation, repeated across his 2025 essays, is that **AI gets you about 70% of the way to a working solution; the last 30% is where things get hard**. The 30% is where logic edge-cases, security boundaries, cross-file invariants, and integration with messy reality live. The dangerous shape is that the 70% feels like 95% — confident syntax, plausible cadence, passing happy-path tests — so teams underestimate the last 30% until production reveals it.

Veracode's 45%-OWASP finding and the slopsquatting CVE surge (Georgia Tech's *Vibe Security Radar* tracked 35 CVEs in March 2026 alone) are concrete instances of the last-30% problem in security terms.

---

## Best practices for *reading* AI-generated code

Reading is the load-bearing skill of 2026. The pre-AI ratio of code-you-wrote to code-you-read has inverted; most lines a developer encounters this year were not authored by any human. The practices below distill the consensus that has formed across Willison, Osmani, Böckeler, Beck, Hermans, and the major studies cited above.

1. **Treat AI code as legacy on arrival.** Michael Feathers' definition of legacy code (*code without tests you understand*) applies the moment the model finishes generating. The reading techniques that work on a 10-year-old codebase — characterization tests, refactoring under tests, beacon-hunting — apply unchanged.

2. **Read with the spec or ticket open.** "You cannot assess requirement fidelity if you are only looking at the code." Keep the original intent visible the entire time. This is non-negotiable in the SDD community for a reason: AI code looks more correct than it is, and only the spec disagrees.

3. **Two-pass reading: top-down then bottom-up.** Read file names, signatures, and diff shape first. Form a one-line *hypothesis* about what the change does — write it down before reading bodies. Then read every changed line, comparing against the hypothesis. Surprises against your hypothesis are the most valuable findings in the review. (Hermans, Ch. 4; reinforced by Code Reading Club practice.)

4. **Treat the AI like an over-confident junior.** Birgitta Böckeler's metaphor (Thoughtworks) is the cleanest mental posture: trust the *eagerness* but verify the *content*. The AI does not know what it does not know. Its tone is independent of its correctness.

5. **Tests are read separately from implementation.** AI often writes tests that mirror the implementation rather than the specification, creating an illusion of correctness when read together. Read the test alone, predict the implementation, then check. Mismatches are findings.

6. **Hunt for beacons; flag generic names.** Beacons (Brooks 1983; Hermans Ch. 4) are names and structural choices that carry real domain meaning. AI tends to produce *generically good* names — `data`, `result`, `processItem` — that look fine and convey nothing. Specifically good names — `unbilled_invoices`, `stripe_charge_response` — are anchors that signal the author understood the domain. Generic names in AI code are a "did anyone here actually think about this?" signal.

7. **Run the code, don't just read it.** The single highest-ROI verification step. AI code often looks right and breaks on empty values, nulls, very large numbers, or unexpected types. Bright Security's 2026 review guide calls this "non-negotiable."

8. **Verify package existence before accepting any import.** The 20%-package-hallucination figure is now a security-audit basic. Cross-check every new dependency against the registry; treat any "fresh" package name with suspicion (the slopsquatting attack relies on registering names the LLM hallucinated).

9. **Externalize beacons as notes.** When you understand *why* a hunk is the way it is, write it down — in your own words, not the AI's. This is the practice Unravel is built around. The note is a durable beacon that survives into future sessions, and the act of writing it is the comprehension test (Ousterhout: "if you can't explain it, you don't understand it").

10. **Pass the modification, deletion, and explanation tests for every hunk you mark done.** "Could I change this safely? Could I argue for removing it? Could I explain it without rereading?" If the answer to any of the three is no, the review is not done — even if every line has been read.

---

## Best practices for *building* with AI

The complement to reading is building. The 2025–2026 consensus is that the value of an AI agent is bounded by the quality of the surrounding engineering system. Six practices show up in nearly every serious account (DORA, Osmani, Beck, Willison, Böckeler):

1. **Write the spec first; treat it as source of truth.** A good spec defines six things: outcomes, scope boundaries, constraints, prior decisions, task breakdown, and verification criteria. Edit the spec when requirements change; regenerate the code. This is the SDD core loop. Tools that codify it (GitHub Spec Kit, AWS Kiro, OpenSpec) ship a four-phase workflow: Specify → Plan → Tasks → Verify.

2. **Use a separate agent to verify.** The most under-used SDD pattern: don't let the implementing agent grade its own output. A second agent (or a second pass with different framing) catches a non-trivial fraction of issues the first agent will defend. This pattern echoes the human practice of code review by someone other than the author.

3. **Anchor in tests.** Beck's "augmented coding" is essentially TDD with an agent in the keyboard seat. The discipline he was already practicing in 1999 turns out to be the discipline that lets an AI move fast without breaking things: write the test first, generate code until it passes, refactor with the test as safety net. Two decades of TDD literature carry over directly.

4. **Keep the codebase modular and well-named.** Böckeler's empirical observation across Thoughtworks engagements: AI assistants work much better in modular codebases with clear context and dependencies. Legacy / entangled systems make the AI misinterpret relationships. The investment in code clarity now pays a *new* dividend in AI productivity.

5. **Maintain an AI-Prompt Playbook and a Cautionary Tales wiki.** The 2026 review guides converge on this: keep a living document of prompts that worked (with the surrounding context), and a separate living document of patterns the team has reviewed and rejected, with explanations. This is the team's externalized memory of how to collaborate with their AI tools — equivalent to a style guide for the human-AI interface.

6. **Build the platform around AI, not around individuals.** DORA's central 2025 finding: AI amplifies the *engineering system*, not the individual. Organizations that treat AI adoption as a tool drop ("everyone gets Copilot") see weaker effects than organizations that invest in the seven DORA AI capabilities (clear AI policy, healthy data ecosystems, AI-accessible internal data, strong version control, working agreements, user-centric focus, quality internal platforms).

---

## Challenges

The challenges are now well-characterized; none of them are theoretical anymore.

| Challenge | What it looks like | Evidence |
|---|---|---|
| **Confidently incomplete code** | Happy path works; edges, security, concurrency don't. | Veracode 45% OWASP; "70% problem" (Osmani). |
| **Package hallucination / slopsquatting** | LLM imports a package that doesn't exist; attacker registers the name. | USENIX Security 2025; Snyk; Georgia Tech *Vibe Security Radar*. |
| **Code-clone debt** | Repeated logic blocks proliferate; refactoring atrophies. | GitClear 2025: 8× clone growth, refactor share halved. |
| **Cognitive offloading & skill atrophy** | Developers stop building durable mental models; juniors never form them. | Anthropic skill study (−17 pts); Microsoft/CMU; "epistemic debt." |
| **Perception vs. reality of speedup** | Teams feel faster, are actually slower on real tasks. | METR 19% slowdown / 24% perceived speedup. |
| **Stability without surrounding rigor** | Throughput rises, change-failure rate rises faster. | DORA 2025: AI ↑ throughput, ↓ stability without controls. |
| **Locally plausible, globally inconsistent** | Each function reads well; cross-file invariants break silently. | Recurring in Willison, Böckeler, Osmani essays. |
| **Tests that mirror the implementation** | Code and tests both wrong, in the same way. | Documented across SDD and review-guide literature. |

---

## Opportunities

The same 2025 evidence shows a clear positive band: when the surrounding practices are right, AI is a step-change improvement.

- **Greenfield and exploratory work.** METR's authors note the same tool that slows experts on familiar code can deliver ~90% speedup on a greenfield project where nobody has expertise yet. This is precisely Karpathy's original "vibe coding" niche, and it is a real win when scoped honestly.
- **Spec-driven first-pass quality.** Early SDD adopters report 3–10× higher first-pass success on non-trivial tasks. The investment is in writing the spec; the payoff is fewer review/rework cycles.
- **AI as code-review co-pilot.** Properly configured AI reviewers catch 70–80% of low-hanging fruit (style, common bugs, obvious security issues), freeing humans to focus on architecture and business logic. This is the inversion of "AI writes, human reviews" — and it scales better.
- **Learning amplifier when used with engagement.** Anthropic's study showed the *high-scoring* developers shared one habit: cognitive engagement — asking follow-up questions, requesting explanations, combining generation with conceptual learning. Used this way, AI compresses the years-to-competence curve dramatically.
- **Refactoring assistance under tests.** The pattern Beck demonstrated with BPlusTree3: with TDD as guardrail, an agent can carry out refactorings — including non-trivial structural ones — faster than a human, without losing correctness.
- **Documentation, spec extraction, and codebase Q&A.** The least controversial wins of 2025. AI consistently helps with explaining unfamiliar code, answering "where does X happen?", and drafting docs from code. These are reading-aid wins, not authorship wins.

---

## Implications for Unravel

Bringing this back to the product:

- The product hypothesis from `REVIEW_CODE.md` — *protect germane cognitive load; reading is the load-bearing skill* — survives 2025–2026 contact. If anything the empirical case has strengthened: METR (over-estimated speedup), GitClear (clone debt), Anthropic (skill atrophy), and DORA (stability cost) all describe failures of the *review* and *comprehension* stage, not failures of the generation stage.
- The four-frame map (vibe coding / vibe engineering / augmented coding / SDD) is useful for Unravel's positioning: **Unravel is the tool you reach for when you are doing vibe engineering or augmented coding on code that an agent produced.** It is not for vibe coding (no review by design) and it is not for SDD's *authoring* phase (that lives in spec-kit-style tools); it picks up where SDD's *verify* phase begins.
- The "treat AI code as legacy on arrival" framing is the cleanest articulation of what Unravel is for: a Feathers-style tool for code that is legacy the moment it lands.
- The externalized-beacon-as-note pattern is now the consensus best practice (Bright Security's "AI-Prompt Playbook" and "Cautionary Tales" wikis are essentially the same instinct at organizational scale). Per-hunk notes in Unravel are an individual-scale version of the same idea.
- The "AI generates, separate agent verifies" SDD pattern hints at a future Unravel role: a human-anchored verification surface that sits opposite the implementer, instead of trying to compete with implementer agents.

---

## References

The list below is structured by category. Items already in `REVIEW_CODE.md` are not duplicated here; those remain the cognitive-science foundation.

### Primary 2025–2026 essays and talks

- **Andrej Karpathy.** *Vibe coding* — original tweet, Feb 2, 2025. [x.com/karpathy/status/1886192184808149383](https://x.com/karpathy/status/1886192184808149383)
- **Simon Willison.** *Not all AI-assisted programming is vibe coding (but vibe coding rocks).* Mar 19, 2025. [simonwillison.net/2025/Mar/19/vibe-coding](https://simonwillison.net/2025/Mar/19/vibe-coding/)
- **Simon Willison.** Vibe-coding tag archive — ongoing notes. [simonwillison.net/tags/vibe-coding/](https://simonwillison.net/tags/vibe-coding/)
- **Simon Willison.** *Vibe engineering* — coined Oct 2025; see "The AI Coding Paradigm Shift" podcast with Heavybit. [heavybit.com/library/podcasts/high-leverage/ep-9-the-ai-coding-paradigm-shift-with-simon-willison](https://www.heavybit.com/library/podcasts/high-leverage/ep-9-the-ai-coding-paradigm-shift-with-simon-willison)
- **Kent Beck.** *Augmented Coding: Beyond the Vibes.* Tidy First, 2025. [tidyfirst.substack.com/p/augmented-coding-beyond-the-vibes](https://tidyfirst.substack.com/p/augmented-coding-beyond-the-vibes)
- **Kent Beck.** *Augmented Coding & Design.* Tidy First, 2025. [tidyfirst.substack.com/p/augmented-coding-and-design](https://tidyfirst.substack.com/p/augmented-coding-and-design)
- **Kent Beck & Martin Fowler.** *Cycles of disruption in the tech industry* — Pragmatic Engineer conversation. [newsletter.pragmaticengineer.com/p/cycles-of-disruption-in-the-tech](https://newsletter.pragmaticengineer.com/p/cycles-of-disruption-in-the-tech)
- **Addy Osmani.** *Code Review in the Age of AI.* [addyo.substack.com/p/code-review-in-the-age-of-ai](https://addyo.substack.com/p/code-review-in-the-age-of-ai)
- **Addy Osmani.** *Beyond Vibe Coding — A Guide to AI-Assisted Development.* [beyond.addy.ie](https://beyond.addy.ie/)
- **Addy Osmani.** *Vibe coding is not the same as AI-assisted engineering.* [medium.com/@addyosmani/vibe-coding-is-not-the-same-as-ai-assisted-engineering-3f81088d5b98](https://medium.com/@addyosmani/vibe-coding-is-not-the-same-as-ai-assisted-engineering-3f81088d5b98)
- **Addy Osmani.** *My LLM coding workflow going into 2026.* [addyosmani.com/blog/ai-coding-workflow/](https://addyosmani.com/blog/ai-coding-workflow/)
- **Addy Osmani.** *Agent-Skills: Production Engineering for AI Coding Agents.* May 2026. [addyosmani.com/blog/future-agentic-coding/](https://addyosmani.com/blog/future-agentic-coding/)
- **Birgitta Böckeler (Thoughtworks).** *AI-assisted coding: Experiences and perspectives.* [thoughtworks.com/insights/podcasts/technology-podcasts/ai-assisted-coding-experiences-perspectives](https://www.thoughtworks.com/insights/podcasts/technology-podcasts/ai-assisted-coding-experiences-perspectives)
- **Birgitta Böckeler.** *Treat your AI assistant like an overconfident junior developer.* ShiftMag interview. [shiftmag.dev/ai-coding-assistance-6758/](https://shiftmag.dev/ai-coding-assistance-6758/)
- **Birgitta Böckeler.** Publications index. [birgitta.info](https://birgitta.info/)
- **Gergely Orosz.** *Learnings from two years of using AI tools for software engineering.* Pragmatic Engineer. [newsletter.pragmaticengineer.com/p/two-years-of-using-ai](https://newsletter.pragmaticengineer.com/p/two-years-of-using-ai)
- **Stack Overflow.** *Building shared coding guidelines for AI (and people too).* Mar 26, 2026. [stackoverflow.blog/2026/03/26/coding-guidelines-for-ai-agents-and-people-too/](https://stackoverflow.blog/2026/03/26/coding-guidelines-for-ai-agents-and-people-too/)

### Empirical studies (2025)

- **METR.** *Measuring the Impact of Early-2025 AI on Experienced Open-Source Developer Productivity.* arXiv:2507.09089. [arxiv.org/pdf/2507.09089](https://arxiv.org/pdf/2507.09089) — the 19%-slower / 24%-perceived-faster RCT.
- **METR explainer.** Sean Goedecke, *METR's AI productivity study is really good.* [seangoedecke.com/impact-of-ai-study](https://www.seangoedecke.com/impact-of-ai-study/)
- **GitClear.** *AI Copilot Code Quality: 2025 Research.* [gitclear.com/ai_assistant_code_quality_2025_research](https://www.gitclear.com/ai_assistant_code_quality_2025_research) — 211M LoC longitudinal study; 8× clone growth, refactoring halved.
- **DORA / Google Cloud.** *2025 State of AI-Assisted Software Development.* [dora.dev/dora-report-2025/](https://dora.dev/dora-report-2025/) · [cloud.google.com/resources/content/2025-dora-ai-assisted-software-development-report](https://cloud.google.com/resources/content/2025-dora-ai-assisted-software-development-report)
- **Anthropic.** Study on AI coding assistance and skill formation — 17-point mastery drop. [infoq.com/news/2026/02/ai-coding-skill-formation/](https://www.infoq.com/news/2026/02/ai-coding-skill-formation/)
- **USENIX Security 2025.** *We Have a Package for You! A Comprehensive Analysis of Package Hallucinations.* [usenix.org/publications/loginonline/we-have-package-you-comprehensive-analysis-package-hallucinations-code](https://www.usenix.org/publications/loginonline/we-have-package-you-comprehensive-analysis-package-hallucinations-code)
- **Veracode.** AI code security analysis — 45% OWASP, 2.74× vulnerability rate. Summarized in [blog.vibecoder.me/security-researchers-ai-code-vulnerability-crisis](https://blog.vibecoder.me/security-researchers-ai-code-vulnerability-crisis)
- **Cloud Security Alliance.** *AI-Generated Code Vulnerability Surge 2026.* [labs.cloudsecurityalliance.org/research/csa-research-note-ai-generated-code-vulnerability-surge-2026/](https://labs.cloudsecurityalliance.org/research/csa-research-note-ai-generated-code-vulnerability-surge-2026/)
- **arXiv 2025.** *A Systematic Literature Review of Code Hallucinations in LLMs.* arXiv:2511.00776. [arxiv.org/pdf/2511.00776](https://arxiv.org/pdf/2511.00776)
- **Microsoft Research & Carnegie Mellon.** Cognitive offloading study on knowledge workers using generative AI (2025).
- **INNOQ.** *Understanding AI Coding Patterns Through Cognitive Load Theory.* Mar 2026. [innoq.com/en/blog/2026/03/ai-cognitive-lens-cognitive-load-theory/](https://www.innoq.com/en/blog/2026/03/ai-cognitive-lens-cognitive-load-theory/)

### Reviewing AI code — practical guides

- **GitHub.** *Review AI-Generated Code* — official Copilot guidance. [docs.github.com/en/copilot/tutorials/review-ai-generated-code](https://docs.github.com/en/copilot/tutorials/review-ai-generated-code)
- **Bright Security.** *5 Best Practices for Reviewing and Approving AI-Generated Code.* [brightsec.com/blog/5-best-practices-for-reviewing-and-approving-ai-generated-code/](https://brightsec.com/blog/5-best-practices-for-reviewing-and-approving-ai-generated-code/)
- **Apiiro.** *10 Best Practices That Will Transform Your Code Review Processes.* [apiiro.com/blog/best-practices-to-transform-your-code-review-process/](https://apiiro.com/blog/best-practices-to-transform-your-code-review-process/)
- **Graphite.** *AI code review implementation and best practices.* [graphite.com/guides/ai-code-review-implementation-best-practices](https://graphite.com/guides/ai-code-review-implementation-best-practices)
- **JavaWorld.** *Code Reviews in the Age of AI: Best Practices for 2026 Teams.* [javaworldmag.com/evolving-code-reviews-with-ai-in-2026/](https://javaworldmag.com/evolving-code-reviews-with-ai-in-2026/)
- **diffray.** *LLM Hallucinations in AI Code Review.* [diffray.ai/blog/llm-hallucinations-code-review/](https://diffray.ai/blog/llm-hallucinations-code-review/)
- **Snyk.** *Slopsquatting: New AI Hallucination Threats & Mitigation Strategies.* [snyk.io/articles/slopsquatting-mitigation-strategies/](https://snyk.io/articles/slopsquatting-mitigation-strategies/)

### Spec-Driven Development

- **GitHub.** *Spec Kit* — official toolkit. [marktechpost.com/2026/05/08/meet-github-spec-kit-an-open-source-toolkit-for-spec-driven-development-with-ai-coding-agents/](https://www.marktechpost.com/2026/05/08/meet-github-spec-kit-an-open-source-toolkit-for-spec-driven-development-with-ai-coding-agents/)
- **Microsoft Developer Blog.** *Diving Into Spec-Driven Development with GitHub Spec Kit.* [developer.microsoft.com/blog/spec-driven-development-spec-kit](https://developer.microsoft.com/blog/spec-driven-development-spec-kit)
- **DeepLearning.AI.** Course: *Spec-Driven Development with Coding Agents.* [deeplearning.ai/courses/spec-driven-development-with-coding-agents](https://www.deeplearning.ai/courses/spec-driven-development-with-coding-agents)
- **Towards Data Science.** *From Vibe Coding to Spec-Driven Development.* [towardsdatascience.com/from-vibe-coding-to-spec-driven-development/](https://towardsdatascience.com/from-vibe-coding-to-spec-driven-development/)
- **BCMS.** *Spec-Driven Development: The Definitive 2026 Guide.* [thebcms.com/blog/spec-driven-development](https://thebcms.com/blog/spec-driven-development)
- **MarkTechPost.** *9 Best AI Tools for Spec-Driven Development in 2026.* [marktechpost.com/2026/05/08/9-best-ai-tools-for-spec-driven-development-in-2026-kiro-bmad-gsd-and-more-compare/](https://www.marktechpost.com/2026/05/08/9-best-ai-tools-for-spec-driven-development-in-2026-kiro-bmad-gsd-and-more-compare/)
- **arXiv 2025.** *Vibe Coding: Toward an AI-Native Paradigm for Semantic and Intent-Driven Programming.* arXiv:2510.17842. [arxiv.org/html/2510.17842v1](https://arxiv.org/html/2510.17842v1)

### Foundations carried over from `REVIEW_CODE.md`

The cognitive-science foundations (Hermans, Sweller, de Groot, Sajaniemi, Soloway & Ehrlich, Brooks) are in `REVIEW_CODE.md` and apply unchanged. The practitioner foundations (Spolsky on reading vs. rewriting; Feathers on legacy code; Ousterhout on complexity; McConnell on review; Fowler on refactoring) also carry over — every one of them turns out to be *more* applicable to AI code, not less.

---

## One-line summary

> The 2025–2026 evidence has settled the argument: AI does not remove the need to read, refactor, and own code — it raises the price of skipping that work. The tools, practices, and frames that minimize that price (vibe engineering, augmented coding, SDD, spec-first review, externalized beacons) are the ones worth investing in. Unravel is one piece of that stack: the reading surface for the code an agent just dropped on your branch.
