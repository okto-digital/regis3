# Clean Code Instructions
## Distilled from "Clean Code" by Robert C. Martin

---

## CORE PHILOSOPHY

The only way to go fast is to keep the code clean. Messy code slows you down instantly.
Leave the code cleaner than you found it (Boy Scout Rule).
Clean code reads like well-written prose - it reveals the designer's intent clearly.

---

## NAMING

### Fundamental Rules
- Names must reveal intent - a name should tell you why it exists, what it does, and how it's used
- If a name requires a comment to explain it, the name is wrong
- Use pronounceable names you can discuss without sounding foolish
- Use searchable names - single-letter names and numeric constants are hard to locate
- The length of a name should correspond to the size of its scope

### What to Avoid
- Disinformation: Don't use names with hidden or misleading meanings
- Noise words: `Data`, `Info`, `Manager`, `Processor` add nothing
- Number series: `a1`, `a2`, `a3` provide no intent
- Hungarian notation and type encodings: Modern IDEs make them unnecessary
- Mental mapping: Readers shouldn't need to translate names in their heads
- Single-letter variables except in tiny loops

### Classes and Functions
- Class names: Use nouns or noun phrases (`Customer`, `WikiPage`, `Account`)
- Method names: Use verbs or verb phrases (`postPayment`, `deletePage`, `save`)
- Accessors, mutators, predicates: Prefix with `get`, `set`, `is`
- Pick one word per concept and use it consistently throughout the codebase
- Use solution domain names (CS terms) when appropriate, problem domain names otherwise

### Context
- Add meaningful context with well-named classes, functions, and namespaces
- Don't add gratuitous context - `GSD` prefix on every class is noise
- Shorter names are better if they are clear

---

## FUNCTIONS

### Size
- Functions should be small - 20 lines is already pushing it
- 2-4 lines is ideal - each function tells a story leading to the next
- Blocks within `if`, `else`, `while` should be one line long (a function call)
- Indent level should not exceed one or two

### Single Responsibility
**FUNCTIONS SHOULD DO ONE THING. THEY SHOULD DO IT WELL. THEY SHOULD DO IT ONLY.**

How to know if it does one thing: If you can extract another function from it with a name that isn't merely a restatement of its implementation, it's doing more than one thing.

### Abstraction Levels
- Statements within a function should all be at the same level of abstraction
- Code should read top-down like a narrative (Stepdown Rule)
- Each function leads to the next at decreasing abstraction levels

### Arguments
- Zero arguments (niladic) is best
- One argument (monadic) is good
- Two arguments (dyadic) is acceptable but harder to understand
- Three arguments (triadic) should be avoided
- More than three requires very special justification

**Argument anti-patterns:**
- Flag arguments (booleans) are ugly - they declare the function does multiple things
- Output arguments are counterintuitive - prefer changing state of owning object
- When function needs 2-3 args, consider wrapping them in a class

**Good monadic forms:**
- Asking a question: `boolean fileExists("file")`
- Transforming input: `InputStream fileOpen("file")`
- Event handlers: `void passwordAttemptFailed(int attempts)`

### Side Effects
Side effects are lies. Don't do hidden things like initializing sessions in a `checkPassword` function.
If temporal coupling exists, make it explicit in the function name.

### Command Query Separation
Functions should either DO something or ANSWER something, not both.
```
// Bad
if (set("username", "bob"))...

// Good  
if (attributeExists("username")) {
    setAttribute("username", "bob");
}
```

### Error Handling
- Prefer exceptions over error codes
- Extract try/catch blocks into their own functions
- Error handling is one thing - a function that handles errors should do nothing else
- If `try` exists in a function, it should be the first word and nothing after catch/finally
- Don't return null - throw exception or return Special Case object
- Don't pass null - consider it a coding error

### DRY (Don't Repeat Yourself)
Duplication is the root of all evil in software. Every duplication represents a missed abstraction opportunity.

---

## COMMENTS

### Philosophy
Comments are failures - we write them because we fail to express ourselves in code.
The only truly good comment is the one you found a way not to write.
Truth is found only in code - comments lie as code evolves and comments don't follow.

### Good Comments (Rare)
- Legal comments (copyright, license)
- Explanation of intent when code cannot express it
- Clarification of obscure library code you can't change
- Warning of consequences
- TODO comments (but use sparingly)
- Javadocs in public APIs

### Bad Comments (Delete Them)
- Redundant comments that say what the code already says
- Misleading comments that don't accurately describe the code
- Mandated comments (not every function needs a javadoc)
- Journal comments (that's what version control is for)
- Noise comments ("Default constructor")
- Position markers (`// Actions /////////`)
- Commented-out code - DELETE IT, version control remembers
- Attribution ("Added by John") - use version control
- HTML in comments

### The Rule
**If you can express it in code, express it in code:**
```
// Bad
// Check if employee is eligible for full benefits
if ((employee.flags & HOURLY_FLAG) && (employee.age > 65))

// Good
if (employee.isEligibleForFullBenefits())
```

---

## FORMATTING

### Purpose
Formatting is about communication. Code formatting affects maintainability and extensibility.

### Vertical Formatting
- Files should typically be 200-500 lines (smaller is better)
- Newspaper metaphor: Name tells story, details increase as you go down
- Vertical openness: Blank lines separate concepts
- Vertical density: Related code should be vertically close
- Vertical distance: Related concepts should be near each other
- Variable declarations: As close to usage as possible
- Instance variables: At the top of the class
- Dependent functions: Caller above callee (top-down)

### Horizontal Formatting
- Lines should not require horizontal scrolling (80-120 chars)
- Horizontal openness: Spaces around operators
- Indentation: Respect hierarchy, never collapse short statements

### Team Rules
A team should agree on formatting rules and everyone follows them.
The code should look like it was written by one person.

---

## OBJECTS AND DATA STRUCTURES

### Data Abstraction
- Hide implementation - expose abstract interfaces
- Don't blindly add getters/setters
- Think about the best way to represent the data an object contains

### The Law of Demeter
A module should not know about the innards of the objects it manipulates.
Method `f` of class `C` should only call methods on:
- `C` itself
- Objects created by `f`
- Objects passed as arguments to `f`
- Objects held in instance variables of `C`

**Avoid train wrecks:**
```
// Bad
output = ctxt.getOptions().getScratchDir().getAbsolutePath();

// Better - but ask yourself if you even need this
String outputDir = ctxt.getAbsolutePathOfScratchDirectory();
```

### Objects vs Data Structures
- Objects: Hide data, expose behavior
- Data structures: Expose data, have no significant behavior
- Hybrid structures that do both are the worst of both worlds

---

## ERROR HANDLING

### Principles
- Write your try-catch-finally statement first
- Use unchecked exceptions (checked exceptions violate Open/Closed Principle)
- Provide context with exceptions - include operation and failure type
- Define exception classes in terms of caller's needs
- Don't return null - throw or return special case
- Don't pass null

### Special Case Pattern
Instead of handling null everywhere:
```
// Create class that encapsulates the special behavior
public class NullEmployee extends Employee {
    public Money getPay() { return Money.ZERO; }
}
```

---

## BOUNDARIES

### Using Third-Party Code
- Wrap third-party APIs to limit dependency spread
- Don't let third-party types leak throughout codebase
- Write learning tests to understand and document third-party behavior
- Learning tests verify that packages work as expected after upgrades

```javascript
// BAD: Third-party type everywhere
const sensors = new Map();
sensors.get(sensorId);

// GOOD: Wrapped in domain-specific class
class Sensors {
  private sensors = new Map();
  getById(id) { return this.sensors.get(id); }
}
```

### Code That Doesn't Exist Yet
- Define interfaces you wish existed
- Use ADAPTER pattern to bridge when real implementation arrives
- Creates testing seam with fakes

### Key Quote
> Depend on something you control, not something you don't control, lest it end up controlling you.

---

## UNIT TESTS

### Three Laws of TDD
1. You may not write production code until you have written a failing unit test
2. You may not write more of a unit test than is sufficient to fail
3. You may not write more production code than is sufficient to pass the test

### Clean Tests
Test code is as important as production code.
Without tests, every change is a potential bug.

**F.I.R.S.T. Principles:**
- **Fast**: Tests should run quickly
- **Independent**: Tests should not depend on each other
- **Repeatable**: Tests should run in any environment
- **Self-Validating**: Tests should have boolean output (pass/fail)
- **Timely**: Write tests just before production code

### One Assert per Test (Guideline)
Minimize asserts per test - single concept per test is the real rule.

### Test Structure: BUILD-OPERATE-CHECK
```javascript
// BUILD: Set up test data
makePages("PageOne", "PageOne.ChildOne", "PageTwo");

// OPERATE: Execute the operation
submitRequest("root", "type:pages");

// CHECK: Verify results
assertResponseIsXML();
assertResponseContains("<name>PageOne</name>");
```

### Domain-Specific Testing Language
Build utility functions that make tests read like specifications:
- `makePageWithContent(name, content)` not raw API calls
- `assertResponseContains(text)` not string parsing
- Let tests express intent, not mechanism

---

## CLASSES

### Organization
1. Public static constants
2. Private static variables
3. Private instance variables
4. Public functions
5. Private utilities called by public functions (stepdown rule)

### Size
- Classes should be small - measured in responsibilities, not lines
- Single Responsibility Principle: A class should have one reason to change
- Class name should describe its responsibility in ~25 words without "if", "and", "or", "but"

### Cohesion
- Classes should have a small number of instance variables
- Each method should manipulate one or more of those variables
- High cohesion: Every variable is used by every method

### Open-Closed Principle
Classes should be open for extension but closed for modification.
Use abstractions to isolate from change.

### Dependency Inversion
Depend on abstractions, not concretions.
High-level modules should not depend on low-level modules.

---

## SMELLS AND HEURISTICS

### Comments
- **C1**: Inappropriate information (belongs in version control, issue tracker)
- **C2**: Obsolete comment (delete or update immediately)
- **C3**: Redundant comment (code explains itself)
- **C4**: Poorly written comment (if writing, do it well)
- **C5**: Commented-out code (delete it)

### Environment
- **E1**: Build requires more than one step (should be single command)
- **E2**: Tests require more than one step (should be single command)

### Functions
- **F1**: Too many arguments (max 3, prefer 0-2)
- **F2**: Output arguments (change owning object instead)
- **F3**: Flag arguments (function does multiple things)
- **F4**: Dead function (delete unused code)

### General
- **G1**: Multiple languages in one source file (minimize)
- **G2**: Obvious behavior is unimplemented (follow principle of least surprise)
- **G3**: Incorrect behavior at boundaries (test all edge cases)
- **G4**: Overridden safeties (don't turn off warnings/failing tests)
- **G5**: Duplication (DRY - the most important rule)
- **G6**: Code at wrong level of abstraction (separate high from low)
- **G7**: Base classes depending on derivatives (base should know nothing of derivatives)
- **G8**: Too much information (minimize interface, hide data)
- **G9**: Dead code (delete unreachable code)
- **G10**: Vertical separation (define things close to where used)
- **G11**: Inconsistency (do similar things the same way)
- **G12**: Clutter (remove unused variables, functions, comments)
- **G13**: Artificial coupling (don't couple things that don't need each other)
- **G14**: Feature envy (methods should use their own class's variables)
- **G15**: Selector arguments (avoid booleans that select behavior)
- **G16**: Obscured intent (make code readable, not clever)
- **G17**: Misplaced responsibility (put code where reader expects it)
- **G18**: Inappropriate static (prefer nonstatic if polymorphism might be needed)
- **G19**: Use explanatory variables (break up calculations with named intermediates)
- **G20**: Function names should say what they do
- **G21**: Understand the algorithm (don't just fiddle until tests pass)
- **G22**: Make logical dependencies physical (explicit over implicit)
- **G23**: Prefer polymorphism to if/else or switch/case
- **G24**: Follow standard conventions (team coding standards)
- **G25**: Replace magic numbers with named constants
- **G26**: Be precise (don't be lazy about decisions)
- **G27**: Structure over convention (enforce with structure when possible)
- **G28**: Encapsulate conditionals (`if (shouldBeDeleted(timer))` not `if (timer.hasExpired() && !timer.isRecurrent())`)
- **G29**: Avoid negative conditionals (`if (buffer.shouldCompact())` not `if (!buffer.shouldNotCompact())`)
- **G30**: Functions should do one thing
- **G31**: Hidden temporal couplings (make time dependencies explicit)
- **G32**: Don't be arbitrary (have a reason for structure)
- **G33**: Encapsulate boundary conditions (put `+1` and `-1` adjustments in one place)
- **G34**: Functions should descend only one level of abstraction
- **G35**: Keep configurable data at high levels (defaults at top, not buried)
- **G36**: Avoid transitive navigation (Law of Demeter - talk to friends, not strangers)

### Names
- **N1**: Choose descriptive names (names are 90% of readability)
- **N2**: Choose names at the appropriate level of abstraction
- **N3**: Use standard nomenclature where possible (patterns, conventions)
- **N4**: Unambiguous names
- **N5**: Use long names for long scopes
- **N6**: Avoid encodings (no Hungarian notation)
- **N7**: Names should describe side effects

### Tests
- **T1**: Insufficient tests (test everything that could break)
- **T2**: Use a coverage tool
- **T3**: Don't skip trivial tests
- **T4**: An ignored test is a question about ambiguity
- **T5**: Test boundary conditions
- **T6**: Exhaustively test near bugs (bugs cluster)
- **T7**: Patterns of failure are revealing
- **T8**: Test coverage patterns can be revealing
- **T9**: Tests should be fast

---

## CONCURRENCY

### Principles
- Keep concurrency-related code separate from other code
- Limit access to shared data (encapsulate severely)
- Use copies of data when possible
- Threads should be as independent as possible
- Know your library's thread-safe collections

### Execution Models to Know
- **Producer-Consumer**: Producers add to queue, consumers take from it
- **Readers-Writers**: Balance read throughput vs write consistency
- **Dining Philosophers**: Competing for limited resources (deadlock risks)

### Testing Threaded Code
- Treat spurious failures as potential threading issues (not cosmic rays)
- Get non-threaded code working first (POJOs)
- Make threaded code pluggable for different configurations
- Run with more threads than processors
- Run on different platforms
- Instrument code with jiggling (random yield/sleep) to expose race conditions

### Key Quote
> Do not ignore system failures as one-offs.

---

## KENT BECK'S 4 RULES OF SIMPLE DESIGN

In priority order:

1. **Runs all the tests** - A system that can't be verified shouldn't be deployed
2. **Contains no duplication** - DRY is the primary enemy of good design
3. **Expresses programmer intent** - Code should clearly communicate purpose
4. **Minimizes classes and methods** - Avoid pointless dogmatism; keep counts low

Rule 1 enables rules 2-4. Once you have tests, you can refactor fearlessly.
Rule 4 is lowest priority - don't create interfaces for every class just because.

---

## SYSTEMS

### Separate Construction from Use
- Startup logic (building objects, wiring dependencies) is a different concern from runtime logic
- Don't scatter construction throughout the application
- Move construction to `main` or dedicated factories

### Dependency Injection
- Objects shouldn't instantiate their own dependencies
- Pass dependencies through constructors or setters
- This supports SRP and enables testing with mocks

### Key Quote
> Use the simplest thing that can possibly work.

---

## SUCCESSIVE REFINEMENT

### Process
1. Get it working (it will be messy)
2. Stop when it works
3. Then, and only then, clean it up

It is not enough for code to work. Working code is often badly broken.
Bad code rots and ferments, becoming an inexorable weight that drags the team down.

### Maintaining Cleanliness
- Clean code is relatively easy to maintain
- If you made a mess 5 minutes ago, clean it now
- Never let the rot get started

---