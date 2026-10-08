# Pi session results reading and JSON docs

**Session ID:** ses_fcb224ab7ffeKfbo0wIfct8tpG
**Created:** 8/24/2026, 8:42:20 PM
**Updated:** 8/24/2026, 8:42:55 PM

---

## User

How can one read / stream a pi session results ? provide some docs ; some stuff regarding its json session file, and more. The files in the CWD have nothing to do with this. So no need to bother reading it.

---

## Assistant (Plan · DeepSeek V4 Flash · 6.6s)

_Thinking:_

The user is asking about "pi session results" — this likely refers to Raspberry Pi, but more likely "pi" refers to some tool. Let me think. "pi session" — could be `pi` a command-line tool? There's "pi" as in Raspberry Pi. But "json session file" suggests a tool that writes JSON session files.

Possible interpretations:
1. Raspberry Pi session — unlikely to have JSON session files.
2. Some ML/AI tool named "pi" — e.g., "pi" from Inflection AI (pi.ai), but that's a chat assistant.
3. "pi" could be `pi` the git alias? No.
4. Maybe they mean "pip"? No, "pi session".
5. Could be an agentic CLI tool. There's "claude", "opencode", "cursor" — but "pi"? 

Wait — the user says "read / stream a pi session results ... its json session file". Maybe this refers to a tool called "pi" — perhaps "pi" is a productivity tool? Let me search the web.

Actually, the question is ambiguous. I should search the web for "pi session json file" or "pi CLI session file". There might be a tool called "pi" that has session files. Given the context of opencode (an agentic coding tool), maybe the user uses another agent tool named "pi". 

Hmm, "pi" might refer to "Pi" from "SillyTavern"? Or "pi" — the Raspberry Pi Imager? 

Could be "pi" as in "Python interpreter"? 

Let me just do a web search. The instruction says I can use websearch. The user asks to "provide some docs; some stuff regarding its json session file, and more."

Since the user says the files in CWD have nothing to do with this, I should not read the CWD. I'm in plan mode, so I can only research and present a plan/answer.

Let me search the web for this.

**Tool: websearch**

**Input:**
```json
{}
```

**Error:**
```
Tool execution aborted
```

---

## User

I mean the pi-coding-agent

---

## Assistant (Plan · DeepSeek V4 Flash · 5.1s)

_Thinking:_

The user is asking about "pi-coding-agent" — how to read/stream pi session results, docs about its JSON session file, and more. They clarified they mean the pi-coding-agent.

I'm in plan mode, so read-only. The user said files in CWD have nothing to do with this, so I shouldn't read the CWD. I should search the web for pi-coding-agent documentation.

Let me search for it.

**Tool: websearch**

**Input:**
```json
{
  "query": "pi-coding-agent session json file read stream documentation"
}
```

**Output:**
```
{
  "search_id": "search_ae5fc5c51a6a6367e0f80f7495ef9f29",
  "results": [
    {
      "url": "https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/json.md",
      "title": "pi/packages/coding-agent/docs/json.md at main · earendil-works/pi · GitHub",
      "publish_date": null,
      "excerpts": [
        "# earendil-works/pi — packages/coding-agent/docs/json.md\n\n- Page: GitHub code file\n- URL: https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/json.md\n- Repository: earendil-works/pi\n- Path: packages/coding-agent/docs/json.md\n- Ref: main\n- Lines: 98\n\n---\n\n# JSON Event Stream Mode\n\n```bash\npi --mode json \"Your prompt\"\n```\n\nOutputs all session events as JSON lines to stdout. Useful for integrating pi into other tools or custom UIs.\n\n## Event Types\n\nWire events use `JsonAgentSessionEvent`. It matches\n[`AgentSessionEvent`](https://github.com/earendil-works/pi-mono/blob/main/packages/coding-agent/src/core/agent-session.ts)\nexcept that streaming message updates omit cumulative snapshots:\n\n```typescript\ntype WithoutPartial<T> = T extends { partial: unknown } ? Omit<T, \"partial\"> : T;\n\ntype JsonAssistantMessageEvent<T> = T extends { type: \"toolcall_start\"; partial: unknown }\n  ? WithoutPartial<T> & { id: string; toolName: string }\n  : WithoutPartial<T>;\n\n...\n\nThe first line is the session header:\n\n```json\n{\"type\":\"session\",\"version\":3,\"id\":\"uuid\",\"timestamp\":\"...\",\"cwd\":\"/path\"}\n```\n\nFollowed by events as they occur:\n\n```json\n{\"type\":\"agent_start\"}\n{\"type\":\"turn_start\"}\n{\"type\":\"message_start\",\"message\":{\"role\":\"assistant\",\"content\":[],...}}\n{\"type\":\"message_update\",\"usage\":{...},\"assistantMessageEvent\":{\"type\":\"text_delta\",\"contentIndex\":0,\"delta\":\"Hello\"}}\n{\"type\":\"message_end\",\"message\":{...}}\n{\"type\":\"turn_end\",\"message\":{...},\"toolResults\":[]}\n{\"type\":\"agent_end\",\"messages\":[...]}\n```\n\n`message_update` records are delta-only. They omit both the cumulative `message` field and\n`assistantMessageEvent.partial` to keep stream size linear. The top-level `usage` field contains\nthe latest cumulative provider-reported usage and may remain zero when a provider only reports\nusage at completion. Use `contentIndex` and `delta` to assemble live text, thinking, or tool-call\narguments if needed.\nA `toolcall_start` event also includes the constant-sized `id` and `toolName`\nfields. `message_end` contains the final authoritative message.\n\n## Example\n\n```bash\npi --mode json \"List files\" 2>/dev/null | jq -c 'select(.type == \"message_end\")'\n```"
      ]
    },
    {
      "url": "https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/index.md",
      "title": "pi/packages/coding-agent/docs/index.md at main · earendil-works/pi · GitHub",
      "publish_date": null,
      "excerpts": [
        "# earendil-works/pi — packages/coding-agent/docs/index.md\n\n- Page: GitHub code file\n- URL: https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/index.md\n- Repository: earendil-works/pi\n- Path: packages/coding-agent/docs/index.md\n- Ref: main\n- Lines: 84\n\n---\n\n# Pi Documentation\n\nPi is a minimal terminal coding harness. It is designed to stay small at the core while being extended through TypeScript extensions, skills, prompt templates, themes, and pi packages.\n\n## Quick start\n\nInstall Pi with npm:\n\n```bash\nnpm install -g --ignore-scripts @earendil-works/pi-coding-agent\n```\n\n`--ignore-scripts` disables dependency lifecycle scripts during install. Pi does not require install scripts for normal npm installs.\n\nOn Linux or macOS, you can also use the installer:\n\n```bash\ncurl -fsSL https://pi.dev/install.sh | sh\n```\n\nTo uninstall pi itself, use npm for curl and npm installs:\n\n```bash\nnpm uninstall -g @earendil-works/pi-coding-agent\n```\n\n...\n\n- Sessions - session management, branching, and tree navigation.\n- Compaction - context compaction and branch summarization.\n\n## Customization\n\n- Extensions - TypeScript modules for tools, commands, events, and custom UI.\n- Skills - Agent Skills for reusable on-demand capabilities.\n- Prompt templates - reusable prompts that expand from slash commands.\n- Themes - built-in and custom terminal themes.\n- Pi packages - bundle and share extensions, skills, prompts, and themes.\n- Custom models - add model entries for supported provider APIs.\n- Custom providers - implement custom APIs and OAuth flows.\n\n## Programmatic usage\n\n- SDK - embed pi in Node.js applications.\n- RPC mode - integrate over stdin/stdout JSONL.\n- JSON event stream mode - print mode with structured events.\n- TUI components - build custom terminal UI for extensions.\n\n## Reference\n\n- Environment variables - Pi process configuration and session metadata available to bash tools."
      ]
    },
    {
      "url": "https://hochej.github.io/pi-mono/coding-agent/session/",
      "title": "Session File Format - pi",
      "publish_date": null,
      "excerpts": [
        "logo\n\npi\n\nSession File Format\n\nSearch\n\nlogo pi\n  Coding Agent\n    + Settings\n    + Sessions Sessions\n        \n        On this page\n        \n        - File Location\n        - Deleting Sessions\n        - Session Version\n        - Source Files\n        - Message Types\n                  \n            * Content Blocks\n            * Base Message Types (from pi-ai)\n            * Extended Message Types (from pi-coding-agent)\n            * AgentMessage Union\n        - Entry Base\n        - Entry Types\n                  \n            * SessionHeader\n            * SessionMessageEntry\n            * ModelChangeEntry\n            * ThinkingLevelChangeEntry\n            * CompactionEntry\n            * BranchSummaryEntry\n            * CustomEntry\n            * CustomMessageEntry\n            * LabelEntry\n            * SessionInfoEntry\n        - Tree Structure\n        - Context Building\n        - Parsing Example\n        - SessionManager API\n                  \n            * Static Creation Methods\n* Static Listing Methods\n            * Instance Methods - Session Management\n            * Instance Methods - Appending (all return entry ID)\n            * Instance Methods - Tree Navigation\n            * Instance Methods - Context & Info\n    + Compaction\n    + Keybindings\n    + RPC Protocol\n    + JSON Mode\n    + Shell Aliases\n    + Tree View\n    + Development\n    + Terminal Setup\n    + Windows\n    + Termux\n* Libraries\n  \n  Libraries\n  \n    + pi-ai\n    + pi-agent-core\n    + pi-tui\n    + pi-web-ui\n* Mom (Slack Bot)\n  \n  Mom (Slack Bot)\n  \n    + Overview\n    + Slack Bot Guide\n    + Sandbox\n    + Events\n    + Artifacts Server\n    + v86\n    + New Architecture\n* Pods (GPU)\n  \n  Pods (GPU)\n  \n    + Overview\n    + Models\n    + Implementation Plan\n    + GPT OSS\n    + Qwen3 Coder\n    + Kimi K2\n    + GML 4.5\n* Contributing\n\nOn this page\n\n* File Location\n* Deleting Sessions\n* Session Version\n* Source Files\n* Message Types\n  \n    + Content Blocks\n    + Base Message Types (from pi-ai)\n+ Extended Message Types (from pi-coding-agent)\n    + AgentMessage Union\n* Entry Base\n* Entry Types\n  \n    + SessionHeader\n    + SessionMessageEntry\n    + ModelChangeEntry\n    + ThinkingLevelChangeEntry\n    + CompactionEntry\n    + BranchSummaryEntry\n    + CustomEntry\n    + CustomMessageEntry\n    + LabelEntry\n    + SessionInfoEntry\n* Tree Structure\n* Context Building\n* Parsing Example\n* SessionManager API\n  \n    + Static Creation Methods\n    + Static Listing Methods\n    + Instance Methods - Session Management\n    + Instance Methods - Appending (all return entry ID)\n    + Instance Methods - Tree Navigation\n    + Instance Methods - Context & Info\n\n# Session File Format ¶\n\nSessions are stored as JSONL (JSON Lines) files. Each line is a JSON object with a `type` field. Session entries form a tree structure via `id` / `parentId` fields, enabling in-place branching without creating new files.\n\n## File Location ¶\n\n```\n ~/.pi/agent/sessions/--<path>--/<timestamp>_<uuid>.jsonl\n```\n\n...\n\nThen messages from `firstKeptEntryId` to compaction\n6. Then messages after compaction\n7. Converts `BranchSummaryEntry` and `CustomMessageEntry` to appropriate message formats\n\n## Parsing Example ¶\n\n```\n import { readFileSync } from \"fs\" ;   const lines = readFileSync ( \"session.jsonl\" , \"utf8\" ). trim (). split ( \"\\n\" );   for ( const line of lines ) {  const entry = JSON . parse ( line );   switch ( entry . type ) {  case \"session\" :  console . log ( `Session v ${ entry . version ?? 1 } :  ${ entry . id } ` );  break ;  case \"message\" :  console . log ( `[ ${ entry . id } ]  ${ entry . message . role } :  ${ JSON . stringify ( entry . message . content ) } ` );  break ;  case \"compaction\" :  console . log ( `[ ${ entry . id } ] Compaction:  ${ entry . tokensBefore }  tokens summarized` );  break ;  case \"branch_summary\" :  console . log ( `[ ${ entry . id } ] Branch from  ${ entry . fromId } ` );  break ;  case \"custom\" :  console . log ( `[ ${ entry . id } ] Custom ( ${ entry .\ncustomType } ):  ${ JSON . stringify ( entry . data ) } ` );  break ;  case \"custom_message\" :  console . log ( `[ ${ entry . id } ] Extension message ( ${ entry . customType } ):  ${ entry . content } ` );  break ;  case \"label\" :  console . log ( `[ ${ entry . id } ] Label \" ${ entry . label } \" on  ${ entry . targetId } ` );  break ;  case \"model_change\" :  console . log ( `[ ${ entry . id } ] Model:  ${ entry . provider } / ${ entry . modelId } ` );  break ;  case \"thinking_level_change\" :  console . log ( `[ ${ entry . id } ] Thinking:  ${ entry . thinkingLevel } ` );  break ;  }  }\n```\n\n## SessionManager API ¶\n\nKey methods for working with sessions programmatically.\n\n### Static Creation Methods ¶\n\n* `SessionManager.create(cwd, sessionDir?)` \\- New session\n* `SessionManager.open(path, sessionDir?)` \\- Open existing session file\n* `SessionManager.continueRecent(cwd, sessionDir?)` \\- Continue most recent or create new\n* `SessionManager.inMemory(cwd?)` \\- No file persistence"
      ]
    },
    {
      "url": "https://hochej.github.io/pi-mono/coding-agent/json",
      "title": "JSON Event Stream Mode - pi",
      "publish_date": null,
      "excerpts": [
        "logo\n\npi\n\nJSON Event Stream Mode\n\nSearch\n\nlogo pi\n  Coding Agent\n    + Settings\n    + Sessions\n    + Compaction\n    + Keybindings\n    + RPC Protocol\n    + JSON Mode JSON Mode\n        \n        On this page\n        \n        - Event Types\n        - Message Types\n        - Output Format\n        - Example\n    + Shell Aliases\n    + Tree View\n    + Development\n    + Terminal Setup\n    + Windows\n    + Termux\n* Libraries\n  \n  Libraries\n  \n    + pi-ai\n    + pi-agent-core\n    + pi-tui\n    + pi-web-ui\n* Mom (Slack Bot)\n  \n  Mom (Slack Bot)\n  \n    + Overview\n    + Slack Bot Guide\n    + Sandbox\n    + Events\n    + Artifacts Server\n    + v86\n    + New Architecture\n* Pods (GPU)\n  \n  Pods (GPU)\n  \n    + Overview\n    + Models\n    + Implementation Plan\n    + GPT OSS\n    + Qwen3 Coder\n    + Kimi K2\n    + GML 4.5\n* Contributing\n\nOn this page\n\n* Event Types\n* Message Types\n* Output Format\n* Example\n\n# JSON Event Stream Mode ¶\n\n```\n pi --mode json \"Your prompt\"\n```\nOutputs all session events as JSON lines to stdout. Useful for integrating pi into other tools or custom UIs.\n\n## Event Types ¶\n\nEvents are defined in [`AgentSessionEvent`](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/agent-session.ts) :\n\n```\n type AgentSessionEvent =  | AgentEvent  | { type : \"auto_compaction_start\" ; reason : \"threshold\" | \"overflow\" }  | { type : \"auto_compaction_end\" ; result : CompactionResult | undefined ; aborted : boolean ; willRetry : boolean ; errorMessage? : string }  | { type : \"auto_retry_start\" ; attempt : number ; maxAttempts : number ; delayMs : number ; errorMessage : string }  | { type : \"auto_retry_end\" ; success : boolean ; attempt : number ; finalError? : string };\n```\n\nBase events from [`AgentEvent`](https://github.com/badlogic/pi-mono/blob/main/packages/agent/src/types.ts) :\n\n```\n\n...\n\n\\- `UserMessage` (line 134)\n\\- `AssistantMessage` (line 140)\n\\- `ToolResultMessage` (line 152)\n\nExtended messages from [`packages/coding-agent/src/core/messages.ts`](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/messages.ts) :\n\\- `BashExecutionMessage` (line 29)\n\\- `CustomMessage` (line 46)\n\\- `BranchSummaryMessage` (line 55)\n\\- `CompactionSummaryMessage` (line 62)\n\n## Output Format ¶\n\nEach line is a JSON object. The first line is the session header:\n\n```\n { \"type\" : \"session\" , \"version\" : 3 , \"id\" : \"uuid\" , \"timestamp\" : \"...\" , \"cwd\" : \"/path\" }\n```\n\nFollowed by events as they occur:\n\n```\n { \"type\" : \"agent_start\" }  { \"type\" : \"turn_start\" }  { \"type\" : \"message_start\" , \"message\" :{ \"role\" : \"assistant\" , \"content\" :[], ... }}  { \"type\" : \"message_update\" , \"message\" :{ ... }, \"assistantMessageEvent\" :{ \"type\" : \"text_delta\" , \"delta\" : \"Hello\" , ... }}  { \"type\" : \"message_end\" , \"message\" :{ ... }}  { \"type\" : \"turn_end\" , \"message\" :{ ...\n}, \"toolResults\" :[]}  { \"type\" : \"agent_end\" , \"messages\" :[ ... ]}\n```\n\n## Example ¶\n\n```\n pi --mode json \"List files\" 2 >/dev/null | jq -c 'select(.type == \"message_end\")'\n```\n\nMIT\n\nMade with [Zensical](https://zensical.org/)"
      ]
    },
    {
      "url": "https://pi.dev/docs/latest/session-format",
      "title": "Session File Format · Docs - Pi Coding Agent",
      "publish_date": null,
      "excerpts": [
        "Documentation\n\nGuides and references for configuring and extending Pi.\n\nNavigation\n\nOn this page\n\nFile Location Deleting Sessions Session Version Source Files Message Types Content Blocks Base Message Types (from pi-ai) Extended Message Types (from pi-coding-agent) AgentMessage Union Entry Base Entry Types SessionHeader SessionMessageEntry ModelChangeEntry ThinkingLevelChangeEntry CompactionEntry BranchSummaryEntry CustomEntry CustomMessageEntry LabelEntry SessionInfoEntry Tree Structure Context Building Parsing Example SessionManager API Static Creation Methods Static Listing Methods Instance Methods - Session Management Instance Methods - Appending (all return entry ID) Instance Methods - Tree Navigation Instance Methods - Context & Info\n\nDocumentation\n\n## Start here\n\nOverview Quickstart Using Pi Providers Security Containerization Settings Keybindings Sessions Compaction\n\n## Customization\n\nExtensions Skills Prompt Templates Themes Pi Packages Custom Models Custom Providers\n## Reference\n\nSession Format\n\n## Programmatic Usage\n\nSDK RPC Mode JSON Event Stream Mode TUI Components\n\n## Platform Setup\n\nWindows Termux on Android tmux Terminal Setup Shell Aliases\n\n## Development\n\nDevelopment\n\nSearch documentation\n\n`⌘ K`\n\n## Documentation\n\n## Start here\n\nOverview Quickstart Using Pi Providers Security Containerization Settings Keybindings Sessions Compaction\n\n## Customization\n\nExtensions Skills Prompt Templates Themes Pi Packages Custom Models Custom Providers\n\n## Reference\n\nSession Format\n\n## Programmatic Usage\n\nSDK RPC Mode JSON Event Stream Mode TUI Components\n\n## Platform Setup\n\nWindows Termux on Android tmux Terminal Setup Shell Aliases\n\n## Development\n\nDevelopment\n\n## On this page\n\n...\n\nFile Location Deleting Sessions Session Version Source Files Message Types Content Blocks Base Message Types (from pi-ai) Extended Message Types (from pi-coding-agent) AgentMessage Union Entry Base Entry Types SessionHeader SessionMessageEntry ModelChangeEntry ThinkingLevelChangeEntry CompactionEntry BranchSummaryEntry CustomEntry CustomMessageEntry LabelEntry SessionInfoEntry Tree Structure Context Building Parsing Example SessionManager API Static Creation Methods Static Listing Methods Instance Methods - Session Management Instance Methods - Appending (all return entry ID) Instance Methods - Tree Navigation Instance Methods - Context & Info\n\n# Session File Format\n\nLatest · [](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/session-format.md) · [](https://github.com/earendil-works/pi/edit/main/packages/coding-agent/docs/session-format.md)\n\nSessions are stored as JSONL (JSON Lines) files. Each line is a JSON object with a `type` field."
      ]
    },
    {
      "url": "https://pi.dev/",
      "title": "Pi Coding Agent",
      "publish_date": null,
      "excerpts": [
        "Four modes: interactive, print/JSON, [RPC, and SDK](https://github.com/earendil-works/pi/tree/main/packages/coding-agent) . See [OpenClaw](https://github.com/OpenClaw/OpenClaw) for a real-world integration.\n\nRead the documentation\n\n## Change the harness, not your workflow\n\nPi - Build a custom workflow extension\n\nPi isn't a sealed product. If you need a command, tool, provider, workflow, or UI tweak, just ask Pi to build it. It will customize itself on the fly.\n\nHave Pi manipulate itself in place, hit `/reload` , and keep going. If you think others will find what you built useful, share it!\n\n## 15+ providers, hundreds of models\n\nPi - Switch models mid-session\n\nAnthropic, OpenAI, Google, Azure, Bedrock, Mistral, Groq, Cerebras, xAI, Hugging Face, Kimi For Coding, MiniMax, NVIDIA, OpenRouter, Ollama, and more. Authenticate via API keys or OAuth.\n\nSwitch models mid-session with `/model` or `Ctrl+L` . Cycle through your favorites with `Ctrl+P` .\n\nAdd custom providers and models via [models.\njson](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs/models.md) or [extensions](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs/custom-provider.md) .\n\n## Tree-structured, shareable history\n\nPi - Branch and share a session tree\n\nSessions are stored as trees. Use `/tree` to navigate to any previous point and continue from there. All branches live in a single file. Filter by message type, label entries as bookmarks.\n\nExport to HTML with `/export` , or upload to a GitHub gist with `/share` and get a shareable URL that renders it. [Example session](https://pi.dev/session/) .\n\n## Context engineering\n\nPi - Load project skills and instructions\n\nPi's [minimal system prompt](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/src/core/system-prompt.ts) and extensibility let you do actual context engineering. Control what goes into the context window and how it's managed.\n\n**AGENTS.\n\n...\n\ncom/earendil-works/pi/tree/main/packages/coding-agent) can inject messages before each turn, filter the message history, implement RAG, or build long-term memory.\n\n## Steer or follow up\n\nPi - Steer a running agent\n\nSubmit messages while the agent works. `Enter` sends a steering message (delivered after current tool, interrupts remaining tools). `Alt+Enter` sends a follow-up (waits until the agent finishes).\n\n## Four modes\n\nPi - Generate a shell script in print mode\n\n**Interactive:** The full TUI experience.\n\n**Print/JSON:** `pi -p \"query\"` for scripts, `--mode json` for event streams.\n\n**RPC:** JSON protocol over stdin/stdout for non-Node integrations. See [docs/rpc.md](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs/rpc.md) .\n\n**SDK:** Embed Pi in your apps. See [OpenClaw](https://github.com/OpenClaw/OpenClaw) for a real-world example.\n\n## Primitives, not features\n\nPi - Install a third-party extension\n\n...\n\nOr install a package that does it your way. See the [50+ examples](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/examples/extensions/) .\n\nBundle extensions, skills, prompts, and themes as packages. Install from npm or git:\n\n`$ pi install npm:@foo/pi-tools  \n$ pi install git:github.com/badlogic/pi-doom`\n\nBrowse packages\n\nand it can play doom\n\n## What we didn't build\n\nPi is aggressively extensible so it doesn't have to dictate your workflow. Features that other tools bake in can be built with [extensions](https://github.com/earendil-works/pi/tree/main/packages/coding-agent) , [skills](https://github.com/earendil-works/pi/tree/main/packages/coding-agent) , or installed from third-party Pi packages . This keeps the core minimal while letting you shape Pi to fit how you work.\n\n[Read the blog post](https://mariozechner.at/posts/2025-11-30-pi-coding-agent/)\n\n* ### No MCP\n  \n  Build CLI tools with READMEs (see [Skills](https://github.\n\n...\n\nFull observability, direct interaction.\n\n## Get involved with Pi\n\n* ### Issues\n  \n  [GitHub](https://github.com/earendil-works/pi/issues) for bugs and features.\n* ### Discord\n  \n  [Community server](https://discord.com/invite/nKXTsAcmbT) for discussion and sharing.\n* ### Documentation\n  \n  Read the documentation or browse the [source on GitHub](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/docs) .\n\n[Earendil Inc.](https://earendil.com/) & Contributors\n\nPress Kit\n\nMIT License\n\n[](https://github.com/earendil-works/pi/tree/main/packages/coding-agent \"GitHub\") [](https://www.npmjs.com/package/@earendil-works/pi-coding-agent \"npm\") [](https://discord.com/invite/3cU7Bz4UPx \"Discord\") [](https://x.com/pidotdev \"X\")\n\n[](https://earendil.com \"Earendil Inc. website\") Auto Light Dark\n\npi.dev domain graciously donated by [exe.dev](https://exe.dev)"
      ]
    },
    {
      "url": "https://clawhub.ai/openlark/skills/pi-coding-agent",
      "title": "Pi Coding Agent — ClawHub",
      "publish_date": null,
      "excerpts": [
        "ClawHub\n\nSkills Plugins Official [Docs](https://docs.openclaw.ai/clawhub/)\n\n`K`\n\nskills / openlark / pi-coding-agent\n\n[Development](https://clawhub.ai/skills?category=development)\n\n[\\#coding](https://clawhub.ai/skills?topic=coding) [\\#typescript](https://clawhub.ai/skills?topic=typescript) [\\#prompt](https://clawhub.ai/skills?topic=prompt)\n\n# Pi Coding Agent\n\nUse Pi Coding Agent (@earendil-works/pi-coding-agent) for AI-assisted programming. Pi is an extensible terminal programming assistant supporting multiple model providers, TypeScript extensions, Skills, Prompt Templates, Themes, and Pi Packages.\n\nRead more\n\nOpenLark @openlark\n\n### Install\n\nCLI npx skills Prompt\n\n$\n\n```\nopenclaw skills install  @openlark/pi-coding-agent\n```\n\nSKILL.md Stats & details\n\nSKILL.md Skill Card Files Versions\n\n# Pi Coding Agent\n\nTerminal programming assistant, customizable with TypeScript extensions + Skills + Prompt Templates + Themes.\n\n## Use Cases\nUse for installation, configuration, model switching, extension development, session management, RPC/SDK integration, and more.\n\n## Installation\n\nbash\n\n```\nnpm install -g --ignore-scripts @earendil-works/pi-coding-agent\nexport ANTHROPIC_API_KEY=sk-ant-...; pi   # or use pi → /login to select a provider\n```\n\n## Basic Usage\n\n`pi` — interactive | `pi -p \"q\"` — script | `pi --mode rpc` — cross-language | `pi -c` — continue | `pi -r` — browse sessions\n\n### Four Modes\n\nInteractive (TUI) / `-p` Print / `--mode json` event stream / `--mode rpc` stdin/stdout JSONL [SDK](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fsdk.md&ownerHandle=openlark) embedding: `import { createAgentSession } from \"@earendil-works/pi-coding-agent\"`\n\n### Models & Shortcuts\n\n15+ providers (Anthropic/OpenAI/Google/...), `/login` OAuth or API key. `/model` (Ctrl+L) to switch, `Ctrl+P` to cycle, `Shift+Tab` to cycle thinking level. [Providers detail](https://clawhub.\n\n...\n\npi update                     # update all\npi config                     # enable/disable\n```\n\nPackage: `package.json` with `\"pi\": { \"extensions/skills/prompts/themes\": [...] }` \\+ keyword `\"pi-package\"` . [Package management](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fpackages.md&ownerHandle=openlark)\n\n## Reference Docs\n\n**Runtime API** : [agent-core](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fagent-core.md&ownerHandle=openlark) — Agent/Tool/Hooks/Event stream **LLM API** : [ai](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fai.md&ownerHandle=openlark) — getModel/streamProxy **Extensions** : [extensions](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fextensions.md&ownerHandle=openlark) — Event system | [tui](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Ftui.md&ownerHandle=openlark) — UI components **Integration** : [sdk](https://clawhub.\n\n...\n\nai/api/v1/skills/pi-coding-agent/file?path=references%2Fskills.md&ownerHandle=openlark) — Skill spec | [packages](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fpackages.md&ownerHandle=openlark) — Package distribution **UI** : [shortcuts](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fshortcuts.md&ownerHandle=openlark) — keyboard shortcuts | [themes](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fthemes.md&ownerHandle=openlark) — 51-token themes | [cli](https://clawhub.ai/api/v1/skills/pi-coding-agent/file?path=references%2Fcli.md&ownerHandle=openlark) — all CLI args\n\n## Notes\n\n* Install with `--ignore-scripts` (Pi does not need lifecycle scripts)\n* Pi Packages have full system access; audit source before installing\n* `PI_SKIP_VERSION_CHECK=1` / `PI_TELEMETRY=0` / `PI_OFFLINE=1`\n* Homepage: [pi.dev](https://pi.dev) | Docs: [pi.dev/docs/latest](https://pi.dev/docs/latest)\n\nRead more\n\n## Related skills"
      ]
    },
    {
      "url": "https://pi.ubitools.com/session-format/",
      "title": "Session File Format | Pi Documentation",
      "publish_date": "2026-06-20",
      "excerpts": [
        "Session File Format Sessions are stored as JSONL (JSON Lines) files. Each line is a JSON object with a type field. Session entries form a tree structure via id / parentId fields, enabling in-place branching without creating new files."
      ]
    },
    {
      "url": "https://www.pi-map.org/guides/rpc-and-json-mode",
      "title": "Pi RPC and JSON Mode — Guides | Pi Map",
      "publish_date": null,
      "excerpts": [
        "Pi RPC and JSON Mode — Guides | Pi Map\nThis guide is the deep-dive companion to the official RPC mode and JSON event stream mode documentation. It pulls the rules from packages/coding-agent/docs/rpc.md, json.md, usage.md, and environment-variables.md into one place, with the practical knobs a CI job or scripted host actually needs."
      ]
    },
    {
      "url": "https://pi-handbook.whatsinfor.me/04-sessions/",
      "title": "Session Management - Pi Coding Agent Handbook",
      "publish_date": null,
      "excerpts": [
        "# Pi will ask if you want to fork it into current directory\n```\n\n## Comparison with Amp's Thread Model ¶\n\n|Feature |Amp |Pi |\n| --- | --- | --- |\n|Thread/Session |Server-side threads with IDs |Local JSONL files with IDs |\n|Handoff |Auto-generates focused prompt for new thread |`/fork` \\-- copies history, you edit the starting message |\n|Thread mentioning |`@thread-id` from anywhere |Not native (read session file manually or build extension) |\n|Restore to point |Hover + Restore button |`/tree` \\-- navigate and continue |\n|Thread map |Visual graph of connected threads |`/tree` with filter modes |\n\nThe main gap: pi doesn't have native cross-session referencing. You can't `@mention` another session from within a session. Workaround: ask pi to read the JSONL file, or build a `/recall <session-id>` extension.\n\nPrevious Extension System Next Multi-Model Freedom\n\nUnofficial community handbook · Not affiliated with Mario Zechner or the pi project · Pi is [MIT licensed](https://github."
      ]
    }
  ],
  "warnings": null,
  "usage": [
    {
      "name": "sku_search",
      "count": 1
    }
  ],
  "session_id": "ses_fcb224ab7ffeKfbo0wIfct8tpG"
}
```

**Tool: websearch**

**Input:**
```json
{
  "query": "pi-coding-agent github CLI coding agent"
}
```

**Output:**
```
{
  "search_id": "search_1cd42835a3ffa90482f47eeaa92c9d13",
  "results": [
    {
      "url": "https://github.com/ccjjllt/pi-agent",
      "title": "GitHub - ccjjllt/pi-agent: AI agent toolkit: unified LLM API, agent loop, TUI, coding agent CLI · GitHub",
      "publish_date": null,
      "excerpts": [
        "# ccjjllt/pi-agent\n\n- Page: GitHub repository\n- URL: https://github.com/ccjjllt/pi-agent\n- Description: AI agent toolkit: unified LLM API, agent loop, TUI, coding agent CLI - ccjjllt/pi-agent\n- Stars: 0\n- Forks: 0\n- License: MIT license\n- Default branch: main\n- Created: 2026-06-27T12:38:29.000Z\n- Commits: 4,845\n\n## Top-level files\n\n- .github/\n- .husky/\n- .pi/\n- packages/\n- scripts/\n- .gitattributes\n- .gitignore\n- .npmrc\n- AGENTS.md\n- CONTRIBUTING.md\n- LICENSE\n- README.md\n- SECURITY.md\n- biome.json\n- package-lock.json\n- package.json\n- pi-test.bat\n- pi-test.ps1\n- pi-test.sh\n- test.sh\n- tsconfig.base.json\n- tsconfig.json\n\n## README.md\n\n[pi logo](https://pi.dev/)\n\n[Discord](https://discord.com/invite/3cU7Bz4UPx) [npm](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)\n\n> New issues and PRs from new contributors are auto-closed by default. Maintainers review auto-closed issues daily. See [CONTRIBUTING.md](https://github.com/ccjjllt/pi-agent/blob/main/CONTRIBUTING.md) .\n> \n>\n# Pi Agent Harness\n\nThis is the home of the Pi agent harness project including our self extensible coding agent.\n\n* **[@earendil-works/pi-coding-agent](https://github.com/ccjjllt/pi-agent/blob/main/packages/coding-agent)** : Interactive coding agent CLI\n* **[@earendil-works/pi-agent-core](https://github.com/ccjjllt/pi-agent/blob/main/packages/agent)** : Agent runtime with tool calling and state management\n* **[@earendil-works/pi-ai](https://github.com/ccjjllt/pi-agent/blob/main/packages/ai)** : Unified multi-provider LLM API (OpenAI, Anthropic, Google, …)\n\nTo learn more about Pi:\n\n* [Visit pi.dev](https://pi.dev/) , the project website with demos\n* [Read the documentation](https://pi.dev/docs/latest) , but you can also ask the agent to explain itself\n\n## All Packages\n\n|Package |Description |\n| --- | --- |\n|**[@earendil-works/pi-ai](https://github.com/ccjjllt/pi-agent/blob/main/packages/ai)** |Unified multi-provider LLM API (OpenAI, Anthropic, Google, etc.) |\n|**[@earendil-works/pi-agent-core](https://github.com/ccjjllt/pi-agent/blob/main/packages/agent)** |Agent runtime with tool calling and state management |\n|**[@earendil-works/pi-coding-agent](https://github.com/ccjjllt/pi-agent/blob/main/packages/coding-agent)** |Interactive coding agent CLI |\n|**[@earendil-works/pi-tui](https://github.com/ccjjllt/pi-agent/blob/main/packages/tui)** |Terminal UI library with differential rendering |\n\nFor Slack/chat automation and workflows see [earendil-works/pi-chat](https://github.com/earendil-works/pi-chat) .\n\n## Permissions & Containerization\n\nPi does not include a built-in permission system for restricting filesystem, process, network, or credential access. By default, it runs with the permissions of the user and process that launched it.\n\nIf you need stronger boundaries, containerize or sandbox Pi. See [packages/coding-agent/docs/containerization.md](https://github.com/ccjjllt/pi-agent/blob/main/packages/coding-agent/docs/containerization.\n\n...\n\nsh         # Run pi from sources (can be run from any directory)\n```\n\n## Supply-chain hardening\n\nWe treat npm dependency changes as reviewed code changes.\n\n* Direct external dependencies are pinned to exact versions. Internal workspace packages remain version-ranged.\n* `.npmrc` sets `save-exact=true` and `min-release-age=2` to avoid same-day dependency releases during npm resolution.\n* `package-lock.json` is the dependency ground truth. Pre-commit blocks accidental lockfile commits unless `PI_ALLOW_LOCKFILE_CHANGE=1` is set.\n* `npm run check` verifies pinned direct deps, native TypeScript import compatibility, and the generated coding-agent shrinkwrap.\n* The published CLI package includes `packages/coding-agent/npm-shrinkwrap.json` , generated from the root lockfile, to pin transitive deps for npm users.\n* Release smoke tests use `npm run release:local` to build, pack, and create isolated npm and Bun installs outside the repo before tagging a release.\n* Local release installs, documented npm installs, and `pi update --self` use `--ignore-scripts` where supported.\n* CI installs with `npm ci --ignore-scripts` , and a scheduled GitHub workflow runs `npm audit --omit=dev` plus `npm audit signatures --omit=dev` .\n* Shrinkwrap generation has an explicit allowlist for dependency lifecycle scripts; new lifecycle-script deps fail checks until reviewed.\n\n## Share your OSS coding agent sessions\n\nIf you use Pi or other coding agents for open source work, please share your sessions.\n\nPublic OSS session data helps improve coding agents with real-world tasks, tool use, failures, and fixes instead of toy benchmarks.\n\nFor the full explanation, see [this post on X](https://x.com/badlogicgames/status/2037811643774652911) .\n\nTo publish sessions, use [`badlogic/pi-share-hf`](https://github.com/badlogic/pi-share-hf) . Read its README.md for setup instructions. All you need is a Hugging Face account, the Hugging Face CLI, and `pi-share-hf` ."
      ]
    },
    {
      "url": "https://github.com/meagerfindings/swamp-cli-agent/pull/7",
      "title": "feat: add pi coding agent provider by shelson · Pull Request #7 · meagerfindings/swamp-cli-agent · GitHub",
      "publish_date": null,
      "excerpts": [
        "# feat: add pi coding agent provider\n\n- Page: GitHub pull request\n- URL: https://github.com/meagerfindings/swamp-cli-agent/pull/7\n- Author: shelson\n- Created: 2026-07-24T20:07:46Z\n- Updated: 2026-07-24T20:07:46Z\n- Repository: meagerfindings/swamp-cli-agent\n- Number: #7\n- Branch: shelson:add-pi-agent -> meagerfindings:main\n- Commits: 4\n- Participants: 1\n\n---\n\n## Summary\n\nAdds a seventh provider, [`pi`](https://github.com/earendil-works/pi-coding-agent) , to the CLI agent invoker — developed and adversarially reviewed end-to-end through the extension itself (the pi provider was validated by having a pi agent review its own implementation via `invokeAndParse` ; findings remediated, re-review clean).\n\n## Changes\n\n* **`buildPiCommand`** : `pi --print --mode json --no-session --model <provider/id> [prompt]` . Auth/model catalogs come from pi's existing config ( `~/.pi/agent/` ); the `provider/id` model form selects provider+model in one flag. `piPath` global arg added.\n\n...\n\njson` to every sandboxed CLI — caught in adversarial review (PI-SAFE-1)\n    + Seatbelt (macOS): both profiles now read-deny `~/.pi` for all providers (static profiles can't be provider-scoped; mirrors `~/.config/amp` ). pi on macOS authenticates via env vars — documented in README (PI-SAFE-2)\n* Version bump → `2026.07.24.1` with `upgrades[]` entry; closed-catalog tests updated; pi fixtures/tests added; README updated\n\n## Testing\n\n* `deno check` clean; `deno test -A` : all pi tests pass (remaining failures are the pre-existing Darwin-assumption baseline on a Linux dev box, unchanged from before this branch)\n* Live end-to-end: swamp model instance with `defaultProvider: pi` ran `invoke` and `invokeAndParse` successfully (tokens/cost/duration captured), then performed the structured adversarial review of this very change — including verifying the argv against `pi --help` and diffing the real `--mode json` stream against the extractors\n\n## Notes for reviewers\n* `readonly` profile is narrower for pi than other providers (no grep/glob without bash) — documented in the `PI_TOOL_ARGS` comment\n* The macOS env-auth limitation (sandboxed pi loses `models.json` / `settings.json` /extensions, not just `auth.json` ) is documented in the README and the `.sb` profile comments\n\n* * *\n\n_The pi agent, reviewing the code that lets it review code:_\n\n[snake eating its own tail](https://camo.githubusercontent.com/225640bae070c42848396bebffb0c4e457981a2abed463c792c9664f4837531b/68747470733a2f2f6d65646961312e67697068792e636f6d2f6d656469612f76312e59326c6b505463354d4749334e6a4578643368764e6e5a354d5441314e33497962335a77616a5278656a63356257557862575632626e5a7a624451306544673564446c6861695a6c634431324d563970626e526c636d35686246396e61575a66596e6c666157516d593351395a772f5a6c484734675358654670757233527a76752f67697068792e676966)\n\n## Timeline\n\n- shelson added 4 commits July 25, 2026 07:25 feat: add pi coding agent provider (v2026.07.24.\n1) … 8bd8c0e - New 'pi' provider: pi --print --mode json --no-session -m <provider/id> - JSONL extractors: text from assistant message content …"
      ]
    },
    {
      "url": "https://deepwiki.com/badlogic/pi-mono/4-pi-coding-agent:-coding-agent-cli",
      "title": "pi-coding-agent: Coding Agent CLI | badlogic/pi-mono | DeepWiki",
      "publish_date": "2026-07-21",
      "excerpts": [
        "The coding agent is built around AgentSession, which orchestrates the agent lifecycle, and supports four operational modes: interactive (full TUI), print (single-shot CLI), RPC (JSON protocol over stdio), and programmatic SDK usage."
      ]
    },
    {
      "url": "https://gist.github.com/schpet/85531b6a05a5d8119e859bdec6b0e0b8/",
      "title": "Pi Coding Agent Setup Guide · GitHub",
      "publish_date": "2026-02-04",
      "excerpts": [
        "# Gist: Pi Coding Agent Setup Guide\n\n- Page: GitHub gist\n- URL: https://gist.github.com/schpet/85531b6a05a5d8119e859bdec6b0e0b8/\n- Author: schpet\n- Gist: 85531b6a05a5d8119e859bdec6b0e0b8\n- Files: 1\n\n## pi-setup-guide.md\n\n# Pi Coding Agent Setup Guide\n\n## 1\\. Install the Agent\n\n```shell\nnpm install -g @mariozechner/pi-coding-agent\n```\n\n## 2\\. Launching Pi\n\n```shell\n# Interactive REPL\npi\n\n# One-off prompt\npi -p \"your prompt here\"\n\n# Continue last session (sessions are tree-structured)\npi -c\n```\n\nUse `pi --list-models` to inspect available models, or pass your own provider/model (e.g., `pi --provider openai --model gpt-4o` ). Export the appropriate API key (e.g., `OPENAI_API_KEY` ) before launching Pi.\n\n## 3\\. Configuration Files\n\n|Purpose |Global Location |Project Location |Notes |\n| --- | --- | --- | --- |\n|CLAUDE/AGENT instructions |`~/.pi/agent/CLAUDE.md` (or `AGENTS.md` ) |parent dirs + cwd |All discovered files are concatenated automatically. |\n|System prompt (replace) |`~/.\n\n...\n\n* Install extensions from npm or git: `pi install npm:package@version` , list via `pi list` , upgrade via `pi update` .\n* Popular picks: `pi-mcp-adapter` , `safe-git` , `pi-cost-dashboard` , `pi-notify` , `checkpoint` , `pi-canvas` .\n\n### Installing Skills with `skill.sh`\n\n`skill.sh` is a universal skill installer for Pi, Claude Code, Goose, and Windsurf.\n\n```shell\n# Install a skill for Pi\nskill.sh install pi https://raw.githubusercontent.com/user/repo/main/skill.md\n```\n\nIt downloads the skill markdown and places it under `~/.pi/agent/skills/` , handling any Pi-specific setup. Reuse the same skill definitions across agents for consistent capabilities.\n\n## 8\\. Helpful Resources\n\n* Pi repo: <https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent>\n* MCP adapter: <https://github.com/nicobailon/pi-mcp-adapter>\n* Awesome Pi agent list: <https://github.com/qualisero/awesome-pi-agent>\n* Blog post introduction: <https://lucumr.pocoo.org/2026/1/31/pi/>"
      ]
    },
    {
      "url": "https://www.bitdoze.com/pi-coding-agent-setup-guide",
      "title": "Pi Coding Agent Setup Guide: Install, Configure Models, and Best Extensions",
      "publish_date": null,
      "excerpts": [
        "<li>LazyPi: one-command setup with 60+ skills and 76 themes</li>\n</ul>\n</ListCheck>\n</Notice>\n\n<YouTubeEmbed\n  url=\"https://www.youtube.com/embed/gfw3jBeLeqs\"\n  label=\"Pi Agent Setup Guide: Top Extensions That Make It Unstoppable\"\n/>\n\n\nIf you are still deciding between coding agents, our [OpenCode setup guide](/opencode-setup-guide/) covers the open-source Claude Code alternative, and the [GitHub Copilot alternatives](/github-copilot-alternatives-2026/) article breaks down what to do after the June 1 pricing change.\n\n## What Pi actually is\n\nPi is a terminal coding agent that reads your codebase, plans changes, edits files, runs shell commands, and iterates on failures. The default installation gives you four tools: read, write, edit, and bash. You add everything else through extensions.\n\nThe extension system is TypeScript. You drop a `.ts` file into `~/.pi/agent/extensions/` and Pi loads it.\n\n...\n\nIf you want a single binary with faster startup, there is [pi_agent_rust](https://github.com/Dicklesworthstone/pi_agent_rust) — 823 GitHub stars, 12ms startup, under 8MB.\n\n```bash\ncurl -fsSL https://raw.githubusercontent.com/Dicklesworthstone/pi_agent_rust/main/install.sh | bash\n```\n\nBoth versions read the same config files, so you can switch between them.\n\n## Pi vs OpenCode vs Claude Code\n\n| Feature | Pi | OpenCode | Claude Code |\n|---------|-----|----------|-------------|\n| **Default tools** | 4 | Full set | Full set |\n| **Extension system** | TypeScript | Config + rules | MCP + hooks |\n| **Model choice** | 20+ built-in + custom | 75+ providers | Anthropic only |\n| **Plan mode** | Via extension | Built-in | No |\n| **Memory** | Via extension | No | No |\n| **MCP support** | Via extension | Built-in | Built-in |\n| **Sub-agents** | Via extension | No | No |\n| **Pricing** | Free (pay API) | Free (pay API) | $20/month + API |\n\nPi starts smaller but grows through extensions.\n\n...\n\nThey do not conflict.\n</Accordion>\n\n## Related articles\n\n- [Top AI GitHub repos](/top-ai-github-repos/) — curated catalog: Pi, OpenCode, OpenClaw, Hermes, skills, gateways\n- [OpenCode vs Pi Agent](/opencode-vs-pi-agent/) — side-by-side comparison\n- [TinyFish: Free Web Search for AI Coding Agents](/tinyfish-ai-agents-web-search/) — detailed TinyFish setup guide\n- [Free Web Search for Coding Agents](/tinyfish-free-search-coding-agents/) — focused setup guide\n- [OpenCode Go: 12 AI Models for $10/Month](/opencode-go-plan/) — cheap models for your agent\n- [Hermes Agent setup guide](/hermes-agent-setup-guide/) — install and configure Hermes\n- [OpenCode setup guide](/opencode-setup-guide/) — install and configure OpenCode\n- [Best cheap models for coding agents](/best-cheap-models-hermes-agent/) — model pricing and benchmarks"
      ]
    },
    {
      "url": "https://dev.to/rosgluk/pi-coding-agent-review-minimal-hackable-ai-coding-cli-4ge8",
      "title": "Pi Coding Agent Review: Minimal, Hackable AI Coding CLI",
      "publish_date": "2026-08-23",
      "excerpts": [
        "to%2Frosgluk%2Fpi-coding-agent-review-minimal-hackable-ai-coding-cli-4ge8) [Share to Mastodon](https://s2f.kytta.dev/?text=https%3A%2F%2Fdev.to%2Frosgluk%2Fpi-coding-agent-review-minimal-hackable-ai-coding-cli-4ge8)\n\nReport Abuse\n\nRost\n\nRost\n\nPosted on Aug 23 • Originally published at [glukhov.org](https://www.glukhov.org/ai-devtools/pi/pi-coding-agent-review/)\n\n# Pi Coding Agent Review: Minimal, Hackable AI Coding CLI\n\n\\# llm \\# ai \\# aicoding \\# dev\n\nPi Coding Agent is a minimal, open-source terminal coding harness that ships with four default tools and leaves most of the behaviour to extensions, skills and your own workflow.\n\nMost AI coding agents are becoming bigger products. They add planning modes, subagents, permission layers, IDE integrations, background workers and increasingly elaborate orchestration around the model.\n\n...\n\nThe Pi documentation describes three patterns:\n\n|Pattern |What is isolated |Where credentials live |When to use it |\n| --- | --- | --- | --- |\n|[Gondolin](https://github.com/earendil-works/gondolin) extension |Built-in tools and `!` commands inside a local Linux micro-VM; the `pi` process stays on the host |Provider auth can remain on the host |Local isolation when you want the TUI and keys on the host |\n|Plain Docker |The whole `pi` process |API keys enter the container unless you proxy them |Simple local isolation of the entire agent |\n|[NVIDIA OpenShell](https://docs.nvidia.com/openshell/about/overview) |The whole `pi` process inside a policy-controlled sandbox |The gateway can keep raw model keys outside the sandbox |Local or remote policy over filesystem, network and credentials |\n\nThe important boundary is not just the source directory.\n\n...\n\nShould you switch from Claude Code? Not automatically. Use Claude Code when its built-in permissions, subagents, hooks and workflow fit the way you already work. Evaluate Pi when you want broader ownership of the harness, provider selection, context construction and tool architecture. For advanced developers, I would not even assume the choice has to be exclusive: Claude Code can remain the polished daily driver while Pi becomes the programmable agent laboratory. After enough customisation, that laboratory may turn into the tool you actually prefer.\n\n###  Licensing and cost\n\nPi itself is open source and released under the MIT licence. That does not make model inference free: your actual cost depends on the model provider, API usage, supported subscription authentication, local compute and any sandbox infrastructure you choose to run.\n\n##  References\n\n* Pi Coding Agent repository: <https://github.com/earendil-works/pi>\n* Pi website and installer: <https://pi.dev>\n* npm package: <https://www.npmjs.com/package/@earendil-works/pi-coding-agent>\n* Gondolin micro-VM: <https://github.com/earendil-works/gondolin>\n* NVIDIA OpenShell: <https://docs.nvidia.com/openshell/about/overview>\n* Licence: MIT\n\n## Top comments (0)\n\nSubscribe\n\npic\n\nPersonal Trusted User\n\nCreate template\n\nTemplates let you quickly answer FAQs or store snippets for re-use.\n\nSubmit Preview Dismiss\n\nCode of Conduct • Report abuse\n\nAre you sure you want to hide this comment? It will become hidden in your post, but will still be visible via the comment's permalink .\n\nHide child comments as well\n\nConfirm\n\nFor further actions, you may consider blocking this person and/or reporting abuse\n\nRost\n\nFollow\n\nSoftware developer with interests in photography, music, self-hosting, AI, LLM, Python and Go\n\n* Location\n  \n  Melbourne, Australia\n* Joined\n  \n  Oct 3, 2025\n\n### More from Rost"
      ]
    },
    {
      "url": "https://pi.dev/docs/latest",
      "title": "Pi Documentation · Docs - Pi Coding Agent",
      "publish_date": null,
      "excerpts": [
        "SDK RPC Mode JSON Event Stream Mode TUI Components\n\n## Platform Setup\n\nWindows Termux on Android tmux Terminal Setup Shell Aliases\n\n## Development\n\nDevelopment\n\n## On this page\n\nQuick start Start here Customization Programmatic usage Reference Platform setup Development\n\nOn this page\n\nQuick start Start here Customization Programmatic usage Reference Platform setup Development\n\n# Pi Documentation\n\nLatest · [](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/index.md) · [](https://github.com/earendil-works/pi/edit/main/packages/coding-agent/docs/index.md)\n\nPi is a minimal terminal coding harness. It is designed to stay small at the core while being extended through TypeScript extensions, skills, prompt templates, themes, and pi packages.\n\n## Quick start\n\nCopied\n\nInstall Pi with npm:\n\n```\nnpm install -g --ignore-scripts @earendil-works/pi-coding-agent\n```\n\n`--ignore-scripts` disables dependency lifecycle scripts during install.\n\n...\n\njs applications.\n* RPC mode \\- integrate over stdin/stdout JSONL.\n* JSON event stream mode \\- print mode with structured events.\n* TUI components \\- build custom terminal UI for extensions.\n\n## Reference\n\nCopied\n\n* Environment variables \\- Pi process configuration and session metadata available to bash tools.\n* Session format \\- JSONL session file format, entry types, and SessionManager API.\n\n## Platform setup\n\nCopied\n\n* Windows\n* Termux on Android\n* tmux\n* Terminal setup\n* Shell aliases\n\n## Development\n\nCopied\n\n* Development \\- local setup, project structure, and debugging.\n\n[Earendil Inc.](https://earendil.com/) & Contributors\n\nPress Kit\n\nMIT License\n\n[](https://github.com/earendil-works/pi/tree/main/packages/coding-agent \"GitHub\") [](https://www.npmjs.com/package/@earendil-works/pi-coding-agent \"npm\") [](https://discord.com/invite/3cU7Bz4UPx \"Discord\") [](https://x.com/pidotdev \"X\")\n\n[](https://earendil.com \"Earendil Inc. website\") Auto Light Dark\n\npi."
      ]
    },
    {
      "url": "https://pi.dev/docs/latest/quickstart",
      "title": "Quickstart · Documentation · Pi",
      "publish_date": null,
      "excerpts": [
        "For one-shot prompts:\n\n```\npi -p \"Summarize this codebase\"\ncat README.md | pi -p \"Summarize this text\"\npi -p @screenshot.png \"What's in this image?\"\n```\n\nUse `--mode json` for JSON event output or `--mode rpc` for process integration.\n\n## Next steps\n\nCopied\n\n* Using Pi \\- interactive mode, slash commands, sessions, context files, and CLI reference.\n* Providers \\- authentication and model setup.\n* Settings \\- global and project configuration.\n* Keybindings \\- shortcuts and customization.\n* Pi Packages \\- install shared extensions, skills, prompts, and themes.\n\nPlatform notes: Windows , Termux , tmux , Terminal setup , Shell aliases .\n\n[Earendil Inc.](https://earendil.com/) & Contributors\n\nPress Kit\n\nMIT License\n\n[](https://github.com/earendil-works/pi/tree/main/packages/coding-agent \"GitHub\") [](https://www.npmjs.com/package/@earendil-works/pi-coding-agent \"npm\") [](https://discord.com/invite/3cU7Bz4UPx \"Discord\") [](https://x.com/pidotdev \"X\")\n\n[](https://earendil.com \"Earendil Inc."
      ]
    },
    {
      "url": "https://mschygulla.github.io/posts/local-pi-coding-agent",
      "title": "Pi Coding Agent with local LLM | Martin's Blog",
      "publish_date": null,
      "excerpts": [
        "Martin's Blog\n\n* Home\n* Posts\n* Tags\n* About\n\n# Pi Coding Agent with local LLM\n\nMay 4, 2026\n\nThis article documents my experiments with the Pi Coding Agent and locally hosted Large Language Models (LLMs) served via llama.cpp on a MacBook. The following steps describe the setup process from start to finish.\n\n## Overview\n\nThe diagram below illustrates how the individual components fit together.\n\n```\n┌─────────────────────┐\n│    Pi Coding Agent  │\n└──────────┬──────────┘\n           │ OpenAI-compatible API\n           ▼\n┌─────────────────────┐\n│     llama.cpp       │\n│   (llama-server)    │\n└──────────┬──────────┘\n           │ loads & runs\n           ▼\n┌─────────────────────┐\n│     LLM Model       │\n│  (GGUF on disk)     │\n\n...\n\ncpp and llama-server Setup\n\n[llama.cpp](https://github.com/ggml-org/llama.cpp) is an open-source project that enables you to run Large Language Models locally on your own hardware, without relying on any cloud service.\n\n### 1\\. Install llama.cpp\n\nThe following command installs llama.cpp via Homebrew, including several command-line utilities such as `llama-cli` and `llama-server` .\n\n```\nbrew install llama.cpp\n```\n\n### 2\\. Download LLM Models and Configure llama-server\n\nCreate the directory structure shown below. Download the LLM model files (see the links in the model table above) and place them in the `models` directory.\n\n```\nai-llama-server/\n ├── models/\n │   ├── gemma-4-26B-A4B-it-UD-Q5_K_M.gguf\n │   └── Qwen3.6-27B-UD-Q4_K_XL.gguf\n │   └── Qwen3.6-35B-A3B-UD-Q4_K_M.gguf\n ├── presets.ini\n └── start.sh\n```\n\n`llama-server` is configured via a presets file that lists the available models along with per-model inference parameters.\n\n...\n\nOpen that URL in your browser to interact with the configured models and verify that everything is working correctly.\n\nllama.cpp web UI\n\n## Pi Coding Agent: Installation and Configuration\n\nThe [Pi Coding Agent](https://github.com/badlogic/pi-coding-agent) (commonly referred to simply as Pi) is a minimalist, open-source AI developer tool that runs directly in your terminal. It automates software development tasks by interacting with your files and project environment.\n\n### 1\\. Install nvm and Node.js LTS\n\n[`nvm`](https://github.com/nvm-sh/nvm) (Node Version Manager) is a utility for installing and managing multiple versions of Node.js on a single machine. Since different projects often depend on specific Node.js versions, `nvm` allows you to switch between them instantly without manually reinstalling software.\n\n```\ncurl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash\n```\n\nOnce the Node Version Manager is installed, install the latest Node.js LTS release:\n\n```\nnvm install --lts\n```\nAfter installation, you may need to restart your terminal session for the newly installed tools to become available. If you encounter any issues, refer to the official [troubleshooting guide](https://github.com/nvm-sh/nvm) .\n\n### 2\\. Install and Configure the Pi Coding Agent\n\n```\nnpm install -g @mariozechner/pi-coding-agent\n```\n\nNext, create a configuration file that registers the models we installed and are serving via llama.cpp.\n\nContents of `~/.pi/agent/models.json` :\n\n```\n{\n \"providers\" : {\n \"llamacpp\" : {\n \"baseUrl\" :  \"http://127.0.0.1:8080/v1\" ,\n \"apiKey\" :  \"noop\" ,\n \"api\" :  \"openai-completions\" ,\n \"models\" : [\n         {\n \"id\" :  \"Gemma-4-26B-A4B\" ,\n \"contextWindow\" :  32000 ,\n \"maxTokens\" :  8192         },\n         {\n \"id\" :  \"Qwen3.6-27B\" ,\n \"contextWindow\" :  32000 ,\n \"maxTokens\" :  8192         },\n         {\n \"id\" :  \"Qwen3.6-35B-A3B\" ,\n \"contextWindow\" :  32000 ,\n \"maxTokens\" :  8192         }\n       ]\n     }\n   }\n }\n```\nLaunch `pi` and run the `/model` command to display the list of configured models. Select one of the available models and verify it is functioning correctly by entering a simple prompt such as `calc 2 + 2` — the agent should respond with the correct result.\n\nPi Coding Agent UI\n\n💡 **Note:** As one might expect, `llama-server` must be running before launching `pi` . 😊\n\n## Pi Coding Agent: Observability Extension\n\nTo gain deeper insight into how a local LLM performs, the `pi-observability` extension is a valuable addition. It replaces the default footer with a live observability bar, exposes a full dashboard command, and prints a tokens-per-second (TPS) summary after each agent run.\n\nFor further details, refer to the official [`pi-observability` documentation](https://github.com/imran-vz/pi-observability) .\n\nInstall the extension via the `pi` package manager:\n\n```\npi install npm:pi-observability\n```\n\n## Conclusion\n\n...\n\n* **Tool integrations** — Pi supports custom tools and extensions beyond `pi-observability` ; the [Pi Coding Agent documentation](https://github.com/badlogic/pi-coding-agent) is a good starting point.\n\nRunning capable LLMs locally is no longer reserved for data centres. With the right hardware and a bit of configuration, a MacBook Pro M4 is a surprisingly effective platform for private, latency-free AI-assisted development.\n\nTags:\n\n* AI\n* Llama.cpp\n* Pi Coding Agent"
      ]
    },
    {
      "url": "https://docs.bswen.com/blog/2026-08-10-pi-coding-agent-tutorial-2026",
      "title": "301 Moved Permanently",
      "publish_date": "2026-08-10",
      "excerpts": [
        "Pi Coding Agent Tutorial 2026: Installation, Setup, Workflow ...\nThis post shows how to install Pi, wire up a model provider, and run a real coding workflow, plus the pitfalls that trip up people coming from Claude Code or OpenCode."
      ]
    }
  ],
  "warnings": null,
  "usage": [
    {
      "name": "sku_search",
      "count": 1
    }
  ],
  "session_id": "ses_fcb224ab7ffeKfbo0wIfct8tpG"
}
```

---

## Assistant (Plan · DeepSeek V4 Flash · 4.4s)

_Thinking:_

I found the docs. Let me fetch the official docs for the session format and JSON mode to provide accurate details. The official docs are at pi.dev/docs/latest/session-format and pi.dev/docs/latest/json (JSON Event Stream Mode). Let me fetch those.

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://pi.dev/docs/latest/session-format",
  "format": "markdown"
}
```

**Output:**
```
Session File Format · Documentation · Pi

[](/)

Copy SVG[Download SVG](/logo-auto.svg)[Press Kit](/press-kit)

[Home](/)[Documentation](/docs/latest)[News](/news)[Packages](/packages)[Models](/models)

[](/)Menu

[Home→](/)[Documentation→](/docs/latest)[News→](/news)[Packages→](/packages)[Models→](/models)

[GitHub](https://github.com/earendil-works/pi/tree/main/packages/coding-agent)[npm](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)[Discord](https://discord.com/invite/3cU7Bz4UPx)[X](https://x.com/pidotdev)

[Earendil Inc.](https://earendil.com "Earendil Inc. website")

[Home→](/)[Documentation→](/docs/latest)[News→](/news)[Packages→](/packages)[Models→](/models)

[GitHub](https://github.com/earendil-works/pi/tree/main/packages/coding-agent)[npm](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)[Discord](https://discord.com/invite/3cU7Bz4UPx)[X](https://x.com/pidotdev)

[Earendil Inc.](https://earendil.com "Earendil Inc. website")

Documentation

Guides and references for configuring and extending Pi.

Navigation

On this page

[File Location](#file-location)[Deleting Sessions](#deleting-sessions)[Session Version](#session-version)[Source Files](#source-files)[Message Types](#message-types)[Content Blocks](#content-blocks)[Base Message Types (from pi-ai)](#base-message-types-from-pi-ai)[Extended Message Types (from pi-coding-agent)](#extended-message-types-from-pi-coding-agent)[AgentMessage Union](#agentmessage-union)[Entry Base](#entry-base)[Entry Types](#entry-types)[SessionHeader](#sessionheader)[SessionMessageEntry](#sessionmessageentry)[ModelChangeEntry](#modelchangeentry)[ThinkingLevelChangeEntry](#thinkinglevelchangeentry)[CompactionEntry](#compactionentry)[BranchSummaryEntry](#branchsummaryentry)[CustomEntry](#customentry)[CustomMessageEntry](#custommessageentry)[LabelEntry](#labelentry)[SessionInfoEntry](#sessioninfoentry)[Tree Structure](#tree-structure)[Context Building](#context-building)[Parsing Example](#parsing-example)[SessionManager API](#sessionmanager-api)[Static Creation Methods](#static-creation-methods)[Static Listing Methods](#static-listing-methods)[Instance Methods - Session Management](#instance-methods-session-management)[Instance Methods - Appending (all return entry ID)](#instance-methods-appending-all-return-entry-id)[Instance Methods - Tree Navigation](#instance-methods-tree-navigation)[Instance Methods - Context & Info](#instance-methods-context-info)

Documentation

## Start here

[Overview](/docs/latest)[Quickstart](/docs/latest/quickstart)[Using Pi](/docs/latest/usage)[Providers](/docs/latest/providers)[Security](/docs/latest/security)[Containerization](/docs/latest/containerization)[Settings](/docs/latest/settings)[Keybindings](/docs/latest/keybindings)[Sessions](/docs/latest/sessions)[Compaction](/docs/latest/compaction)

## Customization

[Extensions](/docs/latest/extensions)[Skills](/docs/latest/skills)[Prompt Templates](/docs/latest/prompt-templates)[Themes](/docs/latest/themes)[Pi Packages](/docs/latest/packages)[Custom Models](/docs/latest/models)[Custom Providers](/docs/latest/custom-provider)

## Reference

[Session Format](/docs/latest/session-format)

## Programmatic Usage

[SDK](/docs/latest/sdk)[RPC Mode](/docs/latest/rpc)[JSON Event Stream Mode](/docs/latest/json)[TUI Components](/docs/latest/tui)

## Platform Setup

[Windows](/docs/latest/windows)[Termux on Android](/docs/latest/termux)[tmux](/docs/latest/tmux)[Terminal Setup](/docs/latest/terminal-setup)[Shell Aliases](/docs/latest/shell-aliases)

## Development

[Development](/docs/latest/development)

Search documentation

⌘ K

## Documentation

## Start here

[Overview](/docs/latest)[Quickstart](/docs/latest/quickstart)[Using Pi](/docs/latest/usage)[Providers](/docs/latest/providers)[Security](/docs/latest/security)[Containerization](/docs/latest/containerization)[Settings](/docs/latest/settings)[Keybindings](/docs/latest/keybindings)[Sessions](/docs/latest/sessions)[Compaction](/docs/latest/compaction)

## Customization

[Extensions](/docs/latest/extensions)[Skills](/docs/latest/skills)[Prompt Templates](/docs/latest/prompt-templates)[Themes](/docs/latest/themes)[Pi Packages](/docs/latest/packages)[Custom Models](/docs/latest/models)[Custom Providers](/docs/latest/custom-provider)

## Reference

[Session Format](/docs/latest/session-format)

## Programmatic Usage

[SDK](/docs/latest/sdk)[RPC Mode](/docs/latest/rpc)[JSON Event Stream Mode](/docs/latest/json)[TUI Components](/docs/latest/tui)

## Platform Setup

[Windows](/docs/latest/windows)[Termux on Android](/docs/latest/termux)[tmux](/docs/latest/tmux)[Terminal Setup](/docs/latest/terminal-setup)[Shell Aliases](/docs/latest/shell-aliases)

## Development

[Development](/docs/latest/development)

## On this page

[File Location](#file-location)[Deleting Sessions](#deleting-sessions)[Session Version](#session-version)[Source Files](#source-files)[Message Types](#message-types)[Content Blocks](#content-blocks)[Base Message Types (from pi-ai)](#base-message-types-from-pi-ai)[Extended Message Types (from pi-coding-agent)](#extended-message-types-from-pi-coding-agent)[AgentMessage Union](#agentmessage-union)[Entry Base](#entry-base)[Entry Types](#entry-types)[SessionHeader](#sessionheader)[SessionMessageEntry](#sessionmessageentry)[ModelChangeEntry](#modelchangeentry)[ThinkingLevelChangeEntry](#thinkinglevelchangeentry)[CompactionEntry](#compactionentry)[BranchSummaryEntry](#branchsummaryentry)[CustomEntry](#customentry)[CustomMessageEntry](#custommessageentry)[LabelEntry](#labelentry)[SessionInfoEntry](#sessioninfoentry)[Tree Structure](#tree-structure)[Context Building](#context-building)[Parsing Example](#parsing-example)[SessionManager API](#sessionmanager-api)[Static Creation Methods](#static-creation-methods)[Static Listing Methods](#static-listing-methods)[Instance Methods - Session Management](#instance-methods-session-management)[Instance Methods - Appending (all return entry ID)](#instance-methods-appending-all-return-entry-id)[Instance Methods - Tree Navigation](#instance-methods-tree-navigation)[Instance Methods - Context & Info](#instance-methods-context-info)

On this page

[File Location](#file-location)[Deleting Sessions](#deleting-sessions)[Session Version](#session-version)[Source Files](#source-files)[Message Types](#message-types)[Content Blocks](#content-blocks)[Base Message Types (from pi-ai)](#base-message-types-from-pi-ai)[Extended Message Types (from pi-coding-agent)](#extended-message-types-from-pi-coding-agent)[AgentMessage Union](#agentmessage-union)[Entry Base](#entry-base)[Entry Types](#entry-types)[SessionHeader](#sessionheader)[SessionMessageEntry](#sessionmessageentry)[ModelChangeEntry](#modelchangeentry)[ThinkingLevelChangeEntry](#thinkinglevelchangeentry)[CompactionEntry](#compactionentry)[BranchSummaryEntry](#branchsummaryentry)[CustomEntry](#customentry)[CustomMessageEntry](#custommessageentry)[LabelEntry](#labelentry)[SessionInfoEntry](#sessioninfoentry)[Tree Structure](#tree-structure)[Context Building](#context-building)[Parsing Example](#parsing-example)[SessionManager API](#sessionmanager-api)[Static Creation Methods](#static-creation-methods)[Static Listing Methods](#static-listing-methods)[Instance Methods - Session Management](#instance-methods-session-management)[Instance Methods - Appending (all return entry ID)](#instance-methods-appending-all-return-entry-id)[Instance Methods - Tree Navigation](#instance-methods-tree-navigation)[Instance Methods - Context & Info](#instance-methods-context-info)

# Session File Format

Latest·[](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/session-format.md)·[](https://github.com/earendil-works/pi/edit/main/packages/coding-agent/docs/session-format.md)

Sessions are stored as JSONL (JSON Lines) files. Each line is a JSON object with a `type` field. Session entries form a tree structure via `id`/`parentId` fields, enabling in-place branching without creating new files.

## File Location

[Copied](#file-location)

```
~/.pi/agent/sessions/--<path>--/<timestamp>_<uuid>.jsonl
```

Where `<path>` is the working directory with `/` replaced by `-`.

## Deleting Sessions

[Copied](#deleting-sessions)

Sessions can be removed by deleting their `.jsonl` files under `~/.pi/agent/sessions/`.

Pi also supports deleting sessions interactively from `/resume` (select a session and press `Ctrl+D`, then confirm). When available, pi uses the `trash` CLI to avoid permanent deletion.

## Session Version

[Copied](#session-version)

Sessions have a version field in the header:

-   **Version 1**: Linear entry sequence (legacy, auto-migrated on load)
-   **Version 2**: Tree structure with `id`/`parentId` linking
-   **Version 3**: Renamed `hookMessage` role to `custom` (extensions unification)

Existing sessions are automatically migrated to the current version (v3) when loaded.

## Source Files

[Copied](#source-files)

Source on GitHub ([pi-mono](https://github.com/earendil-works/pi-mono)):

-   [`packages/coding-agent/src/core/session-manager.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/coding-agent/src/core/session-manager.ts) - Session entry types and SessionManager
-   [`packages/coding-agent/src/core/messages.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/coding-agent/src/core/messages.ts) - Extended message types (BashExecutionMessage, CustomMessage, etc.)
-   [`packages/ai/src/types.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/ai/src/types.ts) - Base message types (UserMessage, AssistantMessage, ToolResultMessage)
-   [`packages/agent/src/types.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/agent/src/types.ts) - AgentMessage union type

For TypeScript definitions in your project, inspect `node_modules/@earendil-works/pi-coding-agent/dist/` and `node_modules/@earendil-works/pi-ai/dist/`.

## Message Types

[Copied](#message-types)

Session entries contain `AgentMessage` objects. Understanding these types is essential for parsing sessions and writing extensions.

### Content Blocks

[Copied](#content-blocks)

Messages contain arrays of typed content blocks:

```typescript
interface TextContent {
  type: "text";
  text: string;
}

interface ImageContent {
  type: "image";
  data: string;      // base64 encoded
  mimeType: string;  // e.g., "image/jpeg", "image/png"
}

interface ThinkingContent {
  type: "thinking";
  thinking: string;
}

interface ToolCall {
  type: "toolCall";
  id: string;
  name: string;
  arguments: Record<string, any>;
}
```

### Base Message Types (from pi-ai)

[Copied](#base-message-types-from-pi-ai)

```typescript
interface UserMessage {
  role: "user";
  content: string | (TextContent | ImageContent)[];
  timestamp: number;  // Unix ms
}

interface AssistantMessage {
  role: "assistant";
  content: (TextContent | ThinkingContent | ToolCall)[];
  api: string;
  provider: string;
  model: string;
  usage: Usage;
  stopReason: "stop" | "length" | "toolUse" | "error" | "aborted";
  errorMessage?: string;
  timestamp: number;
}

interface ToolResultMessage {
  role: "toolResult";
  toolCallId: string;
  toolName: string;
  content: (TextContent | ImageContent)[];
  details?: any;      // Tool-specific metadata
  usage?: Usage;      // Nested LLM work performed by the tool
  isError: boolean;
  timestamp: number;
}

interface Usage {
  input: number;
  output: number;
  cacheRead: number;
  cacheWrite: number;
  totalTokens: number;
  cost: {
    input: number;
    output: number;
    cacheRead: number;
    cacheWrite: number;
    total: number;
  };
}
```

The exported pi-ai `StopReason` type also includes `"pending"`, but that value is reserved for partial messages in streaming events. Terminal `done`/`error` messages replace it with a completion reason before pi persists the assistant message, so `"pending"` should never appear in session JSONL.

### Extended Message Types (from pi-coding-agent)

[Copied](#extended-message-types-from-pi-coding-agent)

```typescript
interface BashExecutionMessage {
  role: "bashExecution";
  command: string;
  output: string;
  exitCode: number | undefined;
  cancelled: boolean;
  truncated: boolean;
  fullOutputPath?: string;
  excludeFromContext?: boolean;  // true for !! prefix commands
  timestamp: number;
}

interface CustomMessage {
  role: "custom";
  customType: string;            // Extension identifier
  content: string | (TextContent | ImageContent)[];
  display: boolean;              // Show in TUI
  details?: any;                 // Extension-specific metadata
  timestamp: number;
}

interface BranchSummaryMessage {
  role: "branchSummary";
  summary: string;
  fromId: string;                // Entry we branched from
  timestamp: number;
}

interface CompactionSummaryMessage {
  role: "compactionSummary";
  summary: string;
  tokensBefore: number;
  timestamp: number;
}
```

### AgentMessage Union

[Copied](#agentmessage-union)

```typescript
type AgentMessage =
  | UserMessage
  | AssistantMessage
  | ToolResultMessage
  | BashExecutionMessage
  | CustomMessage
  | BranchSummaryMessage
  | CompactionSummaryMessage;
```

## Entry Base

[Copied](#entry-base)

All entries (except `SessionHeader`) extend `SessionEntryBase`:

```typescript
interface SessionEntryBase {
  type: string;
  id: string;           // 8-char hex ID
  parentId: string | null;  // Parent entry ID (null for first entry)
  timestamp: string;    // ISO timestamp
}
```

## Entry Types

[Copied](#entry-types)

### SessionHeader

[Copied](#sessionheader)

First line of the file. Metadata only, not part of the tree (no `id`/`parentId`).

```json
{"type":"session","version":3,"id":"uuid","timestamp":"2024-12-03T14:00:00.000Z","cwd":"/path/to/project"}
```

For sessions with a parent (created via `/fork`, `/clone`, or `newSession({ parentSession })`):

```json
{"type":"session","version":3,"id":"uuid","timestamp":"2024-12-03T14:00:00.000Z","cwd":"/path/to/project","parentSession":"/path/to/original/session.jsonl"}
```

### SessionMessageEntry

[Copied](#sessionmessageentry)

A message in the conversation. The `message` field contains an `AgentMessage`.

```json
{"type":"message","id":"a1b2c3d4","parentId":"prev1234","timestamp":"2024-12-03T14:00:01.000Z","message":{"role":"user","content":"Hello"}}
{"type":"message","id":"b2c3d4e5","parentId":"a1b2c3d4","timestamp":"2024-12-03T14:00:02.000Z","message":{"role":"assistant","content":[{"type":"text","text":"Hi!"}],"provider":"anthropic","model":"claude-sonnet-4-5","usage":{...},"stopReason":"stop"}}
{"type":"message","id":"c3d4e5f6","parentId":"b2c3d4e5","timestamp":"2024-12-03T14:00:03.000Z","message":{"role":"toolResult","toolCallId":"call_123","toolName":"bash","content":[{"type":"text","text":"output"}],"isError":false}}
```

### ModelChangeEntry

[Copied](#modelchangeentry)

Emitted when the user switches models mid-session.

```json
{"type":"model_change","id":"d4e5f6g7","parentId":"c3d4e5f6","timestamp":"2024-12-03T14:05:00.000Z","provider":"openai","modelId":"gpt-4o"}
```

### ThinkingLevelChangeEntry

[Copied](#thinkinglevelchangeentry)

Emitted when the user changes the thinking/reasoning level.

```json
{"type":"thinking_level_change","id":"e5f6g7h8","parentId":"d4e5f6g7","timestamp":"2024-12-03T14:06:00.000Z","thinkingLevel":"high"}
```

### CompactionEntry

[Copied](#compactionentry)

Created when context is compacted. Stores a summary of earlier messages.

```json
{"type":"compaction","id":"f6g7h8i9","parentId":"e5f6g7h8","timestamp":"2024-12-03T14:10:00.000Z","summary":"User discussed X, Y, Z...","firstKeptEntryId":"c3d4e5f6","tokensBefore":50000}
```

Newer harness-generated compactions embed the retained post-compaction context directly on the entry, instead of `firstKeptEntryId`:

```json
{"type":"compaction","id":"f6g7h8i9","parentId":"e5f6g7h8","timestamp":"2024-12-03T14:10:00.000Z","summary":"User discussed X, Y, Z...","tokensBefore":50000,"retainedTail":[{"role":"user","content":"latest request"},{"role":"assistant","content":[{"type":"text","text":"latest reply"}],"provider":"anthropic","model":"claude-sonnet-4-5","usage":{...},"stopReason":"stop"}]}
```

Optional fields:

-   `usage`: LLM usage from generating the summary; included in session token and cost totals
-   `retainedTail`: Materialized `AgentMessage[]` kept after compaction. This is optional only for backward compatibility with older sessions. Newer harness-generated compactions include it so we can rebuild context from this checkpoint without walking older entries before the compaction entry.
-   `details`: Implementation-specific data (e.g., `{ readFiles: string[], modifiedFiles: string[] }` for default, or custom data for extensions)
-   `fromHook`: `true` if generated by an extension, `false`/`undefined` if pi-generated (legacy field name)
-   `firstKeptEntryId`: for compatibility with old entry format.

### BranchSummaryEntry

[Copied](#branchsummaryentry)

Created when switching branches via `/tree` with an LLM generated summary of the left branch up to the common ancestor. Captures context from the abandoned path.

```json
{"type":"branch_summary","id":"g7h8i9j0","parentId":"a1b2c3d4","timestamp":"2024-12-03T14:15:00.000Z","fromId":"f6g7h8i9","summary":"Branch explored approach A..."}
```

Optional fields:

-   `usage`: LLM usage from generating the summary; included in session token and cost totals
-   `details`: File tracking data (`{ readFiles: string[], modifiedFiles: string[] }`) for default, or custom data for extensions
-   `fromHook`: `true` if generated by an extension, `false`/`undefined` if pi-generated (legacy field name)

### CustomEntry

[Copied](#customentry)

Extension state persistence. Does NOT participate in LLM context.

```json
{"type":"custom","id":"h8i9j0k1","parentId":"g7h8i9j0","timestamp":"2024-12-03T14:20:00.000Z","customType":"my-extension","data":{"count":42}}
```

Use `customType` to identify your extension's entries on reload. Interactive mode can render custom entries via `pi.registerEntryRenderer(customType, renderer)`, but they still do not participate in LLM context.

### CustomMessageEntry

[Copied](#custommessageentry)

Extension-injected messages that DO participate in LLM context.

```json
{"type":"custom_message","id":"i9j0k1l2","parentId":"h8i9j0k1","timestamp":"2024-12-03T14:25:00.000Z","customType":"my-extension","content":"Injected context...","display":true}
```

Fields:

-   `content`: String or `(TextContent | ImageContent)[]` (same as UserMessage)
-   `display`: `true` = show in TUI with distinct styling, `false` = hidden
-   `details`: Optional extension-specific metadata (not sent to LLM)

### LabelEntry

[Copied](#labelentry)

User-defined bookmark/marker on an entry.

```json
{"type":"label","id":"j0k1l2m3","parentId":"i9j0k1l2","timestamp":"2024-12-03T14:30:00.000Z","targetId":"a1b2c3d4","label":"checkpoint-1"}
```

Set `label` to `undefined` to clear a label.

### SessionInfoEntry

[Copied](#sessioninfoentry)

Session metadata (e.g., user-defined display name). Set via `/name`, `--name` / `-n`, or `pi.setSessionName()` in extensions.

```json
{"type":"session_info","id":"k1l2m3n4","parentId":"j0k1l2m3","timestamp":"2024-12-03T14:35:00.000Z","name":"Refactor auth module"}
```

The session name is displayed in the session selector (`/resume`) instead of the first message when set.

## Tree Structure

[Copied](#tree-structure)

Entries form a tree:

-   First entry has `parentId: null`
-   Each subsequent entry points to its parent via `parentId`
-   Branching creates new children from an earlier entry
-   The "leaf" is the current position in the tree

```
[user msg] ─── [assistant] ─── [user msg] ─── [assistant] ─┬─ [user msg] ← current leaf
                                                            │
                                                            └─ [branch_summary] ─── [user msg] ← alternate branch
```

## Context Building

[Copied](#context-building)

`buildContextEntries()` walks from the current leaf to the root, producing the active entry list while honoring compaction:

1.  Collects all entries on the path
2.  If a `CompactionEntry` is on the path:
    -   Includes the compaction entry first
    -   If `retainedTail` is present, it acts as a self-contained checkpoint and entries after the compaction are included
    -   Otherwise entries from `firstKeptEntryId` to the compaction are included
    -   Then entries after compaction are included
3.  Preserves non-message entries in the selected range so interactive mode can render them

`buildSessionContext()` builds on that entry list to produce the message list for the LLM:

1.  Extracts current model and thinking level settings from the full path
2.  Converts selected entries to messages:
    -   `message` -> stored `AgentMessage`
    -   `compaction` -> `compactionSummary` plus `retainedTail` when present
    -   `branch_summary` -> `branchSummary`
    -   `custom_message` -> `CustomMessage`
    -   `custom` -> no context message

This makes newer compactions act like self-contained checkpoints. `retainedTail` is optional only so older sessions that only store `firstKeptEntryId` continue to load correctly.

## Parsing Example

[Copied](#parsing-example)

```typescript
import { readFileSync } from "fs";

const lines = readFileSync("session.jsonl", "utf8").trim().split("\n");

for (const line of lines) {
  const entry = JSON.parse(line);

  switch (entry.type) {
    case "session":
      console.log(`Session v${entry.version ?? 1}: ${entry.id}`);
      break;
    case "message":
      console.log(`[${entry.id}] ${entry.message.role}: ${JSON.stringify(entry.message.content)}`);
      break;
    case "compaction":
      console.log(`[${entry.id}] Compaction: ${entry.tokensBefore} tokens summarized`);
      break;
    case "branch_summary":
      console.log(`[${entry.id}] Branch from ${entry.fromId}`);
      break;
    case "custom":
      console.log(`[${entry.id}] Custom (${entry.customType}): ${JSON.stringify(entry.data)}`);
      break;
    case "custom_message":
      console.log(`[${entry.id}] Extension message (${entry.customType}): ${entry.content}`);
      break;
    case "label":
      console.log(`[${entry.id}] Label "${entry.label}" on ${entry.targetId}`);
      break;
    case "model_change":
      console.log(`[${entry.id}] Model: ${entry.provider}/${entry.modelId}`);
      break;
    case "thinking_level_change":
      console.log(`[${entry.id}] Thinking: ${entry.thinkingLevel}`);
      break;
  }
}
```

## SessionManager API

[Copied](#sessionmanager-api)

Key methods for working with sessions programmatically.

### Static Creation Methods

[Copied](#static-creation-methods)

-   `SessionManager.create(cwd, sessionDir?)` - New session
-   `SessionManager.open(path, sessionDir?)` - Open existing session file
-   `SessionManager.continueRecent(cwd, sessionDir?)` - Continue most recent or create new
-   `SessionManager.inMemory(cwd?)` - No file persistence
-   `SessionManager.forkFrom(sourcePath, targetCwd, sessionDir?)` - Fork session from another project

### Static Listing Methods

[Copied](#static-listing-methods)

-   `SessionManager.list(cwd, sessionDir?, onProgress?)` - List sessions for a directory
-   `SessionManager.listAll(onProgress?)` - List all sessions across all projects

### Instance Methods - Session Management

[Copied](#instance-methods-session-management)

-   `newSession(options?)` - Start a new session (options: `{ parentSession?: string }`)
-   `setSessionFile(path)` - Switch to a different session file
-   `createBranchedSession(leafId)` - Extract branch to new session file

### Instance Methods - Appending (all return entry ID)

[Copied](#instance-methods-appending-all-return-entry-id)

-   `appendMessage(message)` - Add message
-   `appendThinkingLevelChange(level)` - Record thinking change
-   `appendModelChange(provider, modelId)` - Record model change
-   `appendCompaction(summary, firstKeptEntryId, tokensBefore, details?, fromHook?)` - Add compaction
-   `appendCustomEntry(customType, data?)` - Extension state (not in context)
-   `appendSessionInfo(name)` - Set session display name
-   `appendCustomMessageEntry(customType, content, display, details?)` - Extension message (in context)
-   `appendLabelChange(targetId, label)` - Set/clear label

### Instance Methods - Tree Navigation

[Copied](#instance-methods-tree-navigation)

-   `getLeafId()` - Current position
-   `getLeafEntry()` - Get current leaf entry
-   `getEntry(id)` - Get entry by ID
-   `getBranch(fromId?)` - Walk from entry to root
-   `getTree()` - Get full tree structure
-   `getChildren(parentId)` - Get direct children
-   `getLabel(id)` - Get label for entry
-   `branch(entryId)` - Move leaf to earlier entry
-   `resetLeaf()` - Reset leaf to null (before any entries)
-   `branchWithSummary(entryId, summary, details?, fromHook?)` - Branch with context summary

### Instance Methods - Context & Info

[Copied](#instance-methods-context-info)

-   `buildContextEntries()` - Get active branch entries with compaction applied
-   `buildSessionContext()` - Get messages, thinkingLevel, and model for LLM
-   `getEntries()` - All entries (excluding header)
-   `getHeader()` - Session header metadata
-   `getSessionName()` - Get display name from latest session\_info entry
-   `getCwd()` - Working directory
-   `getSessionDir()` - Session storage directory
-   `getSessionId()` - Session UUID
-   `getSessionFile()` - Session file path (undefined for in-memory)
-   `isPersisted()` - Whether session is saved to disk

[Earendil Inc.](https://earendil.com/) & Contributors

[Press Kit](/press-kit)

MIT License

[](https://github.com/earendil-works/pi/tree/main/packages/coding-agent "GitHub")[](https://www.npmjs.com/package/@earendil-works/pi-coding-agent "npm")[](https://discord.com/invite/3cU7Bz4UPx "Discord")[](https://x.com/pidotdev "X")

[](https://earendil.com "Earendil Inc. website")AutoLightDark

pi.dev domain graciously donated by [exe.dev](https://exe.dev)
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://pi.dev/docs/latest/json",
  "format": "markdown"
}
```

**Output:**
```
JSON Event Stream Mode · Documentation · Pi

[](/)

Copy SVG[Download SVG](/logo-auto.svg)[Press Kit](/press-kit)

[Home](/)[Documentation](/docs/latest)[News](/news)[Packages](/packages)[Models](/models)

[](/)Menu

[Home→](/)[Documentation→](/docs/latest)[News→](/news)[Packages→](/packages)[Models→](/models)

[GitHub](https://github.com/earendil-works/pi/tree/main/packages/coding-agent)[npm](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)[Discord](https://discord.com/invite/3cU7Bz4UPx)[X](https://x.com/pidotdev)

[Earendil Inc.](https://earendil.com "Earendil Inc. website")

[Home→](/)[Documentation→](/docs/latest)[News→](/news)[Packages→](/packages)[Models→](/models)

[GitHub](https://github.com/earendil-works/pi/tree/main/packages/coding-agent)[npm](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)[Discord](https://discord.com/invite/3cU7Bz4UPx)[X](https://x.com/pidotdev)

[Earendil Inc.](https://earendil.com "Earendil Inc. website")

Documentation

Guides and references for configuring and extending Pi.

Navigation

On this page

[Event Types](#event-types)[Message Types](#message-types)[Output Format](#output-format)[Example](#example)

Documentation

## Start here

[Overview](/docs/latest)[Quickstart](/docs/latest/quickstart)[Using Pi](/docs/latest/usage)[Providers](/docs/latest/providers)[Security](/docs/latest/security)[Containerization](/docs/latest/containerization)[Settings](/docs/latest/settings)[Keybindings](/docs/latest/keybindings)[Sessions](/docs/latest/sessions)[Compaction](/docs/latest/compaction)

## Customization

[Extensions](/docs/latest/extensions)[Skills](/docs/latest/skills)[Prompt Templates](/docs/latest/prompt-templates)[Themes](/docs/latest/themes)[Pi Packages](/docs/latest/packages)[Custom Models](/docs/latest/models)[Custom Providers](/docs/latest/custom-provider)

## Reference

[Session Format](/docs/latest/session-format)

## Programmatic Usage

[SDK](/docs/latest/sdk)[RPC Mode](/docs/latest/rpc)[JSON Event Stream Mode](/docs/latest/json)[TUI Components](/docs/latest/tui)

## Platform Setup

[Windows](/docs/latest/windows)[Termux on Android](/docs/latest/termux)[tmux](/docs/latest/tmux)[Terminal Setup](/docs/latest/terminal-setup)[Shell Aliases](/docs/latest/shell-aliases)

## Development

[Development](/docs/latest/development)

Search documentation

⌘ K

## Documentation

## Start here

[Overview](/docs/latest)[Quickstart](/docs/latest/quickstart)[Using Pi](/docs/latest/usage)[Providers](/docs/latest/providers)[Security](/docs/latest/security)[Containerization](/docs/latest/containerization)[Settings](/docs/latest/settings)[Keybindings](/docs/latest/keybindings)[Sessions](/docs/latest/sessions)[Compaction](/docs/latest/compaction)

## Customization

[Extensions](/docs/latest/extensions)[Skills](/docs/latest/skills)[Prompt Templates](/docs/latest/prompt-templates)[Themes](/docs/latest/themes)[Pi Packages](/docs/latest/packages)[Custom Models](/docs/latest/models)[Custom Providers](/docs/latest/custom-provider)

## Reference

[Session Format](/docs/latest/session-format)

## Programmatic Usage

[SDK](/docs/latest/sdk)[RPC Mode](/docs/latest/rpc)[JSON Event Stream Mode](/docs/latest/json)[TUI Components](/docs/latest/tui)

## Platform Setup

[Windows](/docs/latest/windows)[Termux on Android](/docs/latest/termux)[tmux](/docs/latest/tmux)[Terminal Setup](/docs/latest/terminal-setup)[Shell Aliases](/docs/latest/shell-aliases)

## Development

[Development](/docs/latest/development)

## On this page

[Event Types](#event-types)[Message Types](#message-types)[Output Format](#output-format)[Example](#example)

On this page

[Event Types](#event-types)[Message Types](#message-types)[Output Format](#output-format)[Example](#example)

# JSON Event Stream Mode

Latest·[](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/json.md)·[](https://github.com/earendil-works/pi/edit/main/packages/coding-agent/docs/json.md)

```bash
pi --mode json "Your prompt"
```

Outputs all session events as JSON lines to stdout. Useful for integrating pi into other tools or custom UIs.

## Event Types

[Copied](#event-types)

Wire events use `JsonAgentSessionEvent`. It matches [`AgentSessionEvent`](https://github.com/earendil-works/pi-mono/blob/main/packages/coding-agent/src/core/agent-session.ts) except that streaming message updates omit cumulative snapshots:

```typescript
type WithoutPartial<T> = T extends { partial: unknown } ? Omit<T, "partial"> : T;

type JsonAssistantMessageEvent<T> = T extends { type: "toolcall_start"; partial: unknown }
  ? WithoutPartial<T> & { id: string; toolName: string }
  : WithoutPartial<T>;

type JsonAgentSessionEvent =
  | Exclude<AgentSessionEvent, { type: "message_update" }>
  | {
      type: "message_update";
      usage: Usage;
      assistantMessageEvent: JsonAssistantMessageEvent<AssistantMessageEvent>;
    };
```

`queue_update` emits the full pending steering and follow-up queues whenever they change. `compaction_start` and `compaction_end` cover both manual and automatic compaction.

Other base events come from [`AgentEvent`](https://github.com/earendil-works/pi-mono/blob/main/packages/agent/src/types.ts):

```typescript
type AgentEvent =
  // Agent lifecycle
  | { type: "agent_start" }
  | { type: "agent_end"; messages: AgentMessage[] }
  // Turn lifecycle
  | { type: "turn_start" }
  | { type: "turn_end"; message: AgentMessage; toolResults: ToolResultMessage[] }
  // Message lifecycle
  | { type: "message_start"; message: AgentMessage }
  | { type: "message_update"; message: AgentMessage; assistantMessageEvent: AssistantMessageEvent }
  | { type: "message_end"; message: AgentMessage }
  // Tool execution
  | { type: "tool_execution_start"; toolCallId: string; toolName: string; args: any }
  | { type: "tool_execution_update"; toolCallId: string; toolName: string; args: any; partialResult: any }
  | { type: "tool_execution_end"; toolCallId: string; toolName: string; result: any; isError: boolean };
```

## Message Types

[Copied](#message-types)

Base messages from [`packages/ai/src/types.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/ai/src/types.ts#L134):

-   `UserMessage` (line 134)
-   `AssistantMessage` (line 140)
-   `ToolResultMessage` (line 152)

Extended messages from [`packages/coding-agent/src/core/messages.ts`](https://github.com/earendil-works/pi-mono/blob/main/packages/coding-agent/src/core/messages.ts#L29):

-   `BashExecutionMessage` (line 29)
-   `CustomMessage` (line 46)
-   `BranchSummaryMessage` (line 55)
-   `CompactionSummaryMessage` (line 62)

## Output Format

[Copied](#output-format)

Each line is a JSON object. The first line is the session header:

```json
{"type":"session","version":3,"id":"uuid","timestamp":"...","cwd":"/path"}
```

Followed by events as they occur:

```json
{"type":"agent_start"}
{"type":"turn_start"}
{"type":"message_start","message":{"role":"assistant","content":[],...}}
{"type":"message_update","usage":{...},"assistantMessageEvent":{"type":"text_delta","contentIndex":0,"delta":"Hello"}}
{"type":"message_end","message":{...}}
{"type":"turn_end","message":{...},"toolResults":[]}
{"type":"agent_end","messages":[...]}
```

`message_update` records are delta-only. They omit both the cumulative `message` field and `assistantMessageEvent.partial` to keep stream size linear. The top-level `usage` field contains the latest cumulative provider-reported usage and may remain zero when a provider only reports usage at completion. Use `contentIndex` and `delta` to assemble live text, thinking, or tool-call arguments if needed. A `toolcall_start` event also includes the constant-sized `id` and `toolName` fields. `message_end` contains the final authoritative message.

## Example

[Copied](#example)

```bash
pi --mode json "List files" 2>/dev/null | jq -c 'select(.type == "message_end")'
```

[Earendil Inc.](https://earendil.com/) & Contributors

[Press Kit](/press-kit)

MIT License

[](https://github.com/earendil-works/pi/tree/main/packages/coding-agent "GitHub")[](https://www.npmjs.com/package/@earendil-works/pi-coding-agent "npm")[](https://discord.com/invite/3cU7Bz4UPx "Discord")[](https://x.com/pidotdev "X")

[](https://earendil.com "Earendil Inc. website")AutoLightDark

pi.dev domain graciously donated by [exe.dev](https://exe.dev)
```

---

## Assistant (Plan · DeepSeek V4 Flash · 9.3s)

_Thinking:_

I now have comprehensive info from the official docs. The user asked: "How can one read / stream a pi session results? provide some docs; some stuff regarding its json session file, and more."

This is an informational request. I'm in plan mode, but this is really a "provide information" question, not a "make changes" task. I can just answer with the docs summary. Plan mode restricts edits; answering questions is fine.

Let me write a concise but complete answer covering:
1. Reading/streaming session results — the JSONL session file on disk
2. Streaming live — `pi --mode json` event stream, and `--mode rpc`
3. Session file format details — location, header, entry types, tree structure
4. Parsing example
5. Links to docs

Keep it reasonably concise but the user asked for "some docs; some stuff regarding its json session file, and more" — so they want a decent amount of detail. I'll structure it with headers. The instruction says keep under 4 lines unless user asks for detail — here the user explicitly asked for docs and detail, so longer is fine.

Pi stores sessions as JSONL files on disk and can also stream them live via `--mode json`. Docs: https://pi.dev/docs/latest (Session Format + JSON Event Stream Mode).

## Streaming live results

```bash
pi --mode json "Your prompt"      # NDJSON event stream on stdout
pi --mode json "List files" 2>/dev/null | jq -c 'select(.type == "message_end")'
```

First line is the session header `{"type":"session","version":3,"id":"uuid","timestamp":"...","cwd":"/path"}`, then events as they occur: `agent_start`, `turn_start`, `message_start`, `message_update`, `message_end`, `turn_end`, `agent_end`, plus `tool_execution_start/update/end`. `message_update` is **delta-only** (`contentIndex` + `delta`) to keep the stream linear — assemble live text from deltas; `message_end` has the final authoritative message. Use `--no-session` if you don't want the file persisted. For bidirectional control there's also `--mode rpc` (JSONL over stdin/stdout).

## Session JSONL file (on disk)

Location: `~/.pi/agent/sessions/--<path>--/<timestamp>_<uuid>.jsonl` where `<path>` is the cwd with `/` → `-`.

Each line is an entry; all except the header carry `{type, id, parentId, timestamp}` forming a tree (branches live in the same file). Entry types:

| type | meaning |
|---|---|
| `session` | header (v3, has `cwd`, optional `parentSession`) |
| `message` | an `AgentMessage` with `role`: `user`, `assistant`, `toolResult`, `bashExecution`, `custom`, `branchSummary`, `compactionSummary` |
| `model_change` / `thinking_level_change` | mid-session switches |
| `compaction` | context compaction (summary, `tokensBefore`, `retainedTail`/`firstKeptEntryId`) |
| `branch_summary` | left-branch context capture when switching via `/tree` |
| `custom` | extension state, not in LLM context |
| `custom_message` | extension-injected message, **in** context |
| `label` | bookmark on `targetId` |
| `session_info` | display name |

Assistant messages include `provider`, `model`, `usage` (tokens + cost), `stopReason`; content is typed blocks (`text`, `thinking`, `image`, `toolCall`). Context building walks from the leaf to root via `parentId`, honoring compactions.

## Parsing

Trivially parse with `jq`/Python/Node — each line is self-contained:

```bash
jq -r 'select(.type=="message") | .message.role + ": " + ((.message.content|tostring))' <session>.jsonl
```

Or use the SDK: `SessionManager.open(path)` / `create(cwd)` / `continueRecent(cwd)`, then `buildSessionContext()`, `getTree()`, `appendMessage()`, etc. from `@earendil-works/pi-coding-agent`. Full entry/API reference: https://pi.dev/docs/latest/session-format

Want me to write you a small reader/parser script (e.g., Python or Node) for these files?

---

