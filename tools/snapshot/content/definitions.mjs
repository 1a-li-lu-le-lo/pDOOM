// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { review } from "./helpers.mjs";

function def(id, term, short, defs, consensus, related = [], see = []) {
  return { id, term, short_definition: short, definitions: defs.map((d) => ({ text: d.text, source_id: d.source ?? null, attribution: d.by, note: d.note ?? "" })), consensus, related_ids: related, see_also_urls: see, ...review("prior", "informational") };
}

export const definitions = [
  def("def-pdoom", "p(DOOM)", "The probability, within a stated horizon and conditional on a stated model specification, that advanced AI leads to outcome O3 (permanent severe disempowerment), O4 (civilizational collapse), O5 (near-extinction), O6 (human extinction), O7 (biospheric catastrophe) or O8 (other irreversible loss). It is a model output, never a direct measurement, and is always shown with its decomposition.", [
    { text: "Probability of doom: an informal term for the subjective probability that AI causes an existential catastrophe; usage varies by speaker.", by: "Common usage (see Wikipedia entry 'P(doom)')", note: "Informal usage rarely states horizon, outcome or conditioning; this observatory always does." },
  ], "contested", ["O3", "O4", "O5", "O6", "O7", "O8"]),
  def("def-agi", "AGI (artificial general intelligence)", "A system that matches or exceeds human performance across most cognitive tasks. Definitions differ on breadth, autonomy and economic framing.", [
    { text: "Highly autonomous systems that outperform humans at most economically valuable work.", by: "OpenAI Charter (2018)", source: "src-openai-charter" },
    { text: "A levelled framework distinguishing performance depth (emerging to superhuman) and generality (narrow to general), with autonomy treated separately.", by: "Morris et al., Levels of AGI (2023)", source: "src-morris-2023-levels-of-agi" },
    { text: "Unaided machines that can accomplish every task better and more cheaply than human workers (high-level machine intelligence).", by: "Grace et al., ESPAI survey wording", source: "src-grace-2024-thousands-of-ai-authors" },
  ], "contested", ["def-asi", "def-frontier-model"]),
  def("def-asi", "ASI (artificial superintelligence)", "An intellect that greatly exceeds the cognitive performance of humans in virtually all domains of interest.", [
    { text: "Any intellect that greatly exceeds the cognitive performance of humans in virtually all domains of interest.", by: "Bostrom, Superintelligence (2014)", source: "src-bostrom-2014-superintelligence" },
  ], "contested", ["def-agi"]),
  def("def-frontier-model", "Frontier model", "A general-purpose model at or near the most capable level available at a given time.", [
    { text: "General-purpose AI models with systemic risk: models with high-impact capabilities, presumed where cumulative training compute exceeds 10^25 floating-point operations.", by: "EU AI Act, Article 51", source: "src-eu-ai-act-oj" },
    { text: "Models covered by developers' frontier safety frameworks, defined by capability thresholds rather than a fixed compute number.", by: "Anthropic RSP; OpenAI Preparedness Framework; Google DeepMind FSF", source: "src-anthropic-rsp" },
  ], "contested", ["def-agi"]),
  def("def-agent", "Agent", "An AI system that pursues goals over multiple steps by observing, planning and acting through tools with limited human involvement.", [
    { text: "Systems that can plan and act autonomously across multiple steps to achieve goals, increasingly with tool use.", by: "International AI Safety Report 2026 (working usage)", source: "src-iasr-2026" },
  ], "contested", ["def-autonomy", "def-tool-use", "def-agentic-harness"]),
  def("def-autonomy", "Autonomy", "The degree to which a system selects and executes actions without human direction or approval, measured here by unattended operation length and approval frequency.", [
    { text: "Levels of autonomy from tool (fully human-controlled) through consultant, collaborator, expert to agent (fully autonomous).", by: "Morris et al., Levels of AGI (2023)", source: "src-morris-2023-levels-of-agi" },
  ], "contested", ["def-agent"]),
  def("def-agentic-harness", "Agentic harness", "The software around a model that turns it into an agent: prompt templates, tool routing, memory, retries, permissions, budgets and logging. Also called scaffold.", [
    { text: "Scaffolding that provides models with tools, memory and control flow; METR reports results 'with METR's scaffold' because harness design changes measured capability.", by: "METR (usage in evaluation reports)", source: "src-metr-2025-measuring-long-tasks" },
  ], "emerging", ["def-harness-of-harnesses", "def-mcp"]),
  def("def-harness-of-harnesses", "Harness of harnesses", "Nested agent systems in which one harness delegates to other harnesses, inheriting or re-scoping authority; raises questions of policy drift, identity ambiguity, duplicate side effects and cancellation propagation.", [
    { text: "Emerging term; closest documented practice is multi-agent orchestration through protocols such as MCP and agent-to-agent delegation.", by: "This observatory (working definition)", note: "No consensus definition exists; the term is defined operationally in research/agents." },
  ], "emerging", ["def-agentic-harness"]),
  def("def-mcp", "MCP (Model Context Protocol)", "An open protocol that connects language-model applications to servers exposing tools, resources and prompts, so that agents can act on external systems in a standard way.", [
    { text: "An open protocol that enables seamless integration between LLM applications and external data sources and tools.", by: "Model Context Protocol specification", source: "src-mcp-specification" },
  ], "consensus", ["def-tool-use", "def-agentic-harness"]),
  def("def-tool-use", "Tool use", "A model invoking external functions, APIs or systems (search, code execution, file access, payments) through structured calls.", [
    { text: "Function calling and tool invocation as defined by MCP tools and vendor APIs.", by: "MCP specification", source: "src-mcp-specification" },
  ], "consensus", ["def-mcp"]),
  def("def-compute", "Compute", "Computational resources used to train or run models, usually measured in floating-point operations; a governance lever because it is detectable, excludable and concentrated.", [
    { text: "Computing power is detectable, excludable, quantifiable and produced via a concentrated supply chain, which makes it a lever for governance.", by: "Sastry et al. (2024)", source: "src-sastry-2024-computing-power-governance" },
  ], "consensus"),
  def("def-capability", "Capability", "What a system can do when suitably elicited, as measured by evaluations and benchmarks; a prerequisite or pressure variable in this observatory, never an outcome.", [
    { text: "Dangerous-capability thresholds defined per domain (biological, chemical, cyber, autonomy) in developer frameworks.", by: "OpenAI Preparedness Framework v2", source: "src-openai-preparedness-v2" },
  ], "contested"),
  def("def-alignment", "Alignment", "The degree to which a system's objectives and behaviour match what its principals intend, including under distribution shift and when unobserved.", [
    { text: "Misalignment: systems pursuing goals that conflict with human intentions; alignment faking: appearing aligned while pursuing other goals.", by: "Greenblatt et al., Alignment faking (2024)", source: "src-anthropic-2024-alignment-faking" },
    { text: "Rogue AIs: the difficulty of controlling agents far more intelligent than humans.", by: "Hendrycks et al. (2023)", source: "src-hendrycks-2023-overview-catastrophic-risks" },
  ], "contested", ["def-control"]),
  def("def-control", "Control", "Safety that does not rely on the model being aligned: protocols that stay safe even if the model tries to subvert them.", [
    { text: "Techniques that maintain safety despite intentional subversion by the model, evaluated with untrusted and trusted models.", by: "Greenblatt et al., AI Control (2023)", source: "src-greenblatt-2023-ai-control" },
  ], "emerging", ["def-alignment", "def-containment"]),
  def("def-containment", "Containment", "Technical and organisational isolation (sandboxes, network egress controls, credential scoping) that limits what a system can affect.", [
    { text: "Least privilege for MCP agent processes: deny access paths that a server does not require.", by: "NSA MCP security guidance (2026)", source: "src-nsa-2026-mcp-security" },
  ], "consensus", ["def-control"]),
  def("def-evaluation", "Evaluation", "A structured test of a system's capabilities or propensities under specified conditions; results depend on scaffold, elicitation and contamination.", [
    { text: "Open framework for LLM evaluations with a published register of tasks.", by: "UK AISI, Inspect", source: "src-aisi-inspect" },
  ], "consensus"),
  def("def-incident", "Incident", "An event in which an AI system's development, use or malfunction led to harm.", [
    { text: "An AI incident is an event, circumstance or series of events where the development, use or malfunction of one or more AI systems directly or indirectly leads to harm.", by: "OECD (2024)", source: "src-oecd-2024-defining-ai-incidents" },
    { text: "Serious incident: an incident leading to death or serious harm to health, serious and irreversible disruption of critical infrastructure, infringement of fundamental rights, or serious harm to property or the environment.", by: "EU AI Act, Article 3(49)", source: "src-eu-ai-act-oj" },
  ], "consensus", ["def-near-miss"]),
  def("def-near-miss", "Near miss", "An event that could plausibly have led to harm but did not, including contained safety failures and controlled-test findings; recorded separately from realised harm.", [
    { text: "AI hazard: an event where an AI system could plausibly lead to an incident.", by: "OECD (2024)", source: "src-oecd-2024-defining-ai-incidents" },
  ], "emerging", ["def-incident"]),
  def("def-catastrophe", "Catastrophe", "Large-scale harm; in this observatory 'catastrophic' outcomes are O4–O8, while O1–O2 are serious harms that leave a recovery path.", [
    { text: "Global catastrophe: death of at least 10% of the world's population within a five-year period (XPT question threshold).", by: "Forecasting Research Institute, XPT", source: "src-fri-2023-xpt" },
  ], "contested", ["O4", "O5", "O6"]),
  def("def-extinction", "Extinction", "No humans survive (outcome O6).", [
    { text: "Human extinction: no humans alive (Metaculus resolution criteria).", by: "Metaculus question 578", source: "src-metaculus-578-human-extinction-2100" },
  ], "consensus", ["O6"]),
  def("def-disempowerment", "Disempowerment", "Loss of humanity's effective control over its future; permanent severe disempowerment is outcome O3.", [
    { text: "Gradual disempowerment: incremental erosion of human influence over the economy, culture and states leading to effectively irreversible loss of influence.", by: "Kulveit et al. (2025)", source: "src-kulveit-2025-gradual-disempowerment" },
    { text: "Similarly permanent and severe disempowerment of the human species (survey wording paired with extinction).", by: "ESPAI surveys", source: "src-grace-2024-thousands-of-ai-authors" },
  ], "contested", ["O3", "def-lock-in"]),
  def("def-lock-in", "Lock-in", "A state of affairs that becomes effectively irreversible, such as permanent value specification or permanent institutional control.", [
    { text: "Existential catastrophe includes unrecoverable dystopia as well as extinction.", by: "Ord, The Precipice (2020)", source: "src-ord-2020-precipice" },
  ], "contested", ["def-recoverability"]),
  def("def-recoverability", "Recoverability", "Whether institutions and societies retain a credible path back from harm; the dividing line between O1–O2 and O3–O8.", [
    { text: "Working definition of this observatory: an outcome is recoverable if, absent further catastrophe, human institutions could plausibly restore autonomy and welfare within generations.", by: "This observatory" },
  ], "emerging", ["def-lock-in"]),
  def("def-systemic-risk", "Systemic risk", "Risk arising from the interaction of many systems, actors and dependencies rather than a single failure.", [
    { text: "Systemic risk from general-purpose AI: risk specific to high-impact capabilities with significant impact on the Union market and society.", by: "EU AI Act, Article 3(65)", source: "src-eu-ai-act-oj" },
    { text: "Societal-scale risks classified by accountability and intent.", by: "Critch & Russell, TASRA (2023)", source: "src-critch-russell-2023-tasra" },
  ], "contested"),
  def("def-malicious-use", "Malicious use", "Deliberate use of AI by humans to cause harm.", [
    { text: "Malicious use: individuals or groups intentionally using AIs to cause harm.", by: "Hendrycks et al. (2023)", source: "src-hendrycks-2023-overview-catastrophic-risks" },
  ], "consensus", ["S4"]),
  def("def-malfunction", "Malfunction", "Harm arising from a system not behaving as intended, without malicious intent.", [
    { text: "Malfunctions, alongside malicious use and systemic risks, form the report's risk categories.", by: "International AI Safety Report 2026", source: "src-iasr-2026" },
  ], "consensus", ["S2"]),
  def("def-loss-of-control", "Loss of control", "A state in which humans cannot reliably monitor, constrain, correct or shut down AI systems.", [
    { text: "Scenarios where models may resist human shutdown or control, now covered by frontier safety framework reviews.", by: "Google DeepMind FSF v3 (2025)", source: "src-gdm-fsf-v3" },
  ], "contested", ["S3"]),
  def("def-epistemic-uncertainty", "Epistemic uncertainty", "Uncertainty from limited knowledge that could in principle be reduced by more evidence; dominant in this observatory.", [
    { text: "Epistemic uncertainty stems from lack of knowledge and is reducible; aleatory uncertainty is inherent randomness.", by: "Der Kiureghian & Ditlevsen (2009)", source: "src-der-kiureghian-2009-aleatory-epistemic" },
  ], "consensus", ["def-aleatory-uncertainty"]),
  def("def-aleatory-uncertainty", "Aleatory uncertainty", "Irreducible randomness in outcomes even with complete knowledge of the process.", [
    { text: "Aleatory uncertainty: inherent randomness that cannot be reduced by more data.", by: "Der Kiureghian & Ditlevsen (2009)", source: "src-der-kiureghian-2009-aleatory-epistemic" },
  ], "consensus", ["def-epistemic-uncertainty"]),
];
