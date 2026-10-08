'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import clsx from 'clsx';
import {
  ArrowDownToLine,
  Bot,
  Brain,
  ChevronDown,
  CircleAlert,
  FilePenLine,
  FileText,
  FolderSearch,
  Loader2,
  Plug,
  ScrollText,
  Search,
  Terminal,
  User,
  Wrench,
} from 'lucide-react';
import type { SessionContentPart, SessionEntry, SessionMessage } from '@/lib/types';
import { formatElapsed, entryKey } from '@/lib/sessions';
import { formatTime } from '@/lib/format';

// ─── render model ────────────────────────────────────────────────────

type ToolStatus = 'running' | 'ok' | 'error' | 'unfinished';

interface ToolCallView {
  id: string;
  name: string;
  args?: string;
  preview?: string;
  status: ToolStatus;
  result?: string;
  at?: string;
  durationMs?: number;
}

type AssistantPart =
  | { kind: 'text'; text: string }
  | { kind: 'thinking'; text: string }
  | { kind: 'tool'; call: ToolCallView };

type Block =
  | { kind: 'user'; key: string; text: string; at?: string }
  | {
      kind: 'assistant';
      key: string;
      parts: AssistantPart[];
      at?: string;
      model?: string;
      stopReason?: string;
      error?: string;
    }
  | { kind: 'bash'; key: string; command: string; output: string; exitCode?: number; at?: string }
  | { kind: 'system'; key: string; text: string; at?: string }
  | { kind: 'meta'; key: string; text: string; detail?: string; at?: string }
  | { kind: 'compaction'; key: string; text: string; at?: string }
  | { kind: 'raw'; key: string; type: string; raw: unknown; at?: string };

interface ToolResult {
  text: string;
  isError: boolean;
}

function partText(parts: SessionContentPart[] | undefined, type: 'text' | 'thinking'): string {
  return (parts ?? [])
    .filter((part) => part.type === type)
    .map((part) => (type === 'text' ? part.text : part.thinking) ?? '')
    .filter(Boolean)
    .join('\n\n');
}

function previewArgs(args: Record<string, unknown> | undefined): string | undefined {
  if (!args) return undefined;
  const stringField = (key: string) => (typeof args[key] === 'string' ? (args[key] as string) : undefined);
  const candidate =
    stringField('command') ??
    stringField('path') ??
    stringField('file_path') ??
    stringField('query') ??
    stringField('pattern') ??
    stringField('url') ??
    stringField('prompt');
  if (!candidate) return undefined;
  const firstLine = candidate.split('\n')[0]?.trim() ?? '';
  return firstLine.length > 160 ? `${firstLine.slice(0, 160)}…` : firstLine;
}

function indentJson(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

/**
 * Turns the raw JSONL entries into the blocks the transcript renders. Tool
 * results are folded into the call that produced them, and a tool call with no
 * result in the newest assistant message is what "running now" means.
 */
function buildBlocks(entries: SessionEntry[], live: boolean): Block[] {
  const results = new Map<string, ToolResult>();
  for (const entry of entries) {
    const message = entry.message;
    if (entry.type !== 'message' || message?.role !== 'toolResult' || !message.toolCallId) continue;
    results.set(message.toolCallId, {
      text: partText(message.content, 'text') || (message.content?.length ? indentJson(message.content) : ''),
      isError: !!message.isError,
    });
  }

  let lastAssistantIndex = -1;
  entries.forEach((entry, index) => {
    if (entry.type === 'message' && entry.message?.role === 'assistant') lastAssistantIndex = index;
  });

  const blocks: Block[] = [];
  entries.forEach((entry, index) => {
    const key = entryKey(entry, index);

    switch (entry.type) {
      case 'session':
        return;
      case 'model_change':
        blocks.push({
          kind: 'meta',
          key,
          text: 'Model',
          detail: `${entry.provider ?? '?'}/${entry.model_id ?? '?'}`,
          at: entry.timestamp,
        });
        return;
      case 'thinking_level_change':
        blocks.push({ kind: 'meta', key, text: 'Thinking level', detail: entry.level, at: entry.timestamp });
        return;
      case 'session_info':
        blocks.push({ kind: 'meta', key, text: 'Session named', detail: entry.name, at: entry.timestamp });
        return;
      case 'label':
        blocks.push({ kind: 'meta', key, text: 'Label', detail: entry.label || '(cleared)', at: entry.timestamp });
        return;
      case 'compaction':
        blocks.push({
          kind: 'compaction',
          key,
          text: entry.summary || 'Context compacted',
          at: entry.timestamp,
        });
        return;
      case 'branch_summary':
        blocks.push({ kind: 'compaction', key, text: entry.summary || 'Branch summary', at: entry.timestamp });
        return;
      case 'custom':
      case 'custom_message':
        blocks.push({
          kind: 'raw',
          key,
          type: entry.custom_type || entry.type,
          raw: entry.raw,
          at: entry.timestamp,
        });
        return;
      case 'message':
        break;
      default:
        blocks.push({ kind: 'raw', key, type: entry.type, raw: entry.raw, at: entry.timestamp });
        return;
    }

    const message = entry.message;
    if (!message) return;

    switch (message.role) {
      case 'user': {
        const text = partText(message.content, 'text');
        if (!text.trim()) return;
        blocks.push({ kind: 'user', key, text, at: entry.timestamp });
        return;
      }
      case 'system': {
        const text = partText(message.content, 'text');
        blocks.push({ kind: 'system', key, text: text || indentJson(entry.raw), at: entry.timestamp });
        return;
      }
      case 'bashExecution': {
        blocks.push({
          kind: 'bash',
          key,
          command: message.command ?? '',
          output: message.output ?? '',
          exitCode: message.exitCode,
          at: entry.timestamp,
        });
        return;
      }
      case 'toolResult': {
        // Rendered inside the tool call that requested it.
        if (message.toolCallId && hasToolCall(entries, message.toolCallId)) return;
        blocks.push({
          kind: 'assistant',
          key,
          at: entry.timestamp,
          parts: [
            {
              kind: 'tool',
              call: {
                id: message.toolCallId ?? key,
                name: message.toolName ?? 'tool',
                status: message.isError ? 'error' : 'ok',
                result: partText(message.content, 'text'),
                at: entry.timestamp,
              },
            },
          ],
        });
        return;
      }
      case 'assistant':
        break;
      default:
        blocks.push({ kind: 'raw', key, type: `${entry.type}:${message.role}`, raw: entry.raw, at: entry.timestamp });
        return;
    }

    const isNewest = index === lastAssistantIndex;
    const parts: AssistantPart[] = [];
    for (const part of message.content ?? []) {
      if (part.type === 'text' && part.text?.trim()) {
        parts.push({ kind: 'text', text: part.text });
      } else if (part.type === 'thinking' && part.thinking?.trim()) {
        parts.push({ kind: 'thinking', text: part.thinking });
      } else if (part.type === 'toolCall') {
        const result = part.id ? results.get(part.id) : undefined;
        const status: ToolStatus = result
          ? result.isError
            ? 'error'
            : 'ok'
          : isNewest && live
            ? 'running'
            : 'unfinished';
        parts.push({
          kind: 'tool',
          call: {
            id: part.id ?? `${key}:${parts.length}`,
            name: part.name ?? 'tool',
            args: part.arguments ? indentJson(part.arguments) : undefined,
            preview: previewArgs(part.arguments),
            status,
            result: result?.text,
            at: entry.timestamp,
          },
        });
      }
    }

    if (parts.length === 0) return;
    blocks.push({
      kind: 'assistant',
      key,
      parts,
      at: entry.timestamp,
      model: message.model,
      stopReason: message.stopReason,
      error: message.errorMessage,
    });
  });

  return blocks;
}

function hasToolCall(entries: SessionEntry[], callId: string): boolean {
  return entries.some(
    (entry) =>
      entry.type === 'message' &&
      entry.message?.role === 'assistant' &&
      (entry.message.content ?? []).some((part) => part.type === 'toolCall' && part.id === callId),
  );
}

// ─── small pieces ────────────────────────────────────────────────────

function toolIcon(name: string) {
  switch (name) {
    case 'bash':
    case 'shell':
      return Terminal;
    case 'read':
      return FileText;
    case 'write':
    case 'edit':
      return FilePenLine;
    case 'grep':
      return Search;
    case 'find':
    case 'ls':
      return FolderSearch;
    default:
      return name.startsWith('mcp') ? Plug : Wrench;
  }
}

function StatusPill({ status, since }: { status: ToolStatus; since?: string }) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (status !== 'running') return;
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [status]);

  if (status === 'running') {
    const started = since ? new Date(since).getTime() : NaN;
    const elapsed = Number.isNaN(started) ? null : Math.max(0, now - started);
    return (
      <span className="chip !border-emerald/30 !text-emerald">
        <Loader2 size={11} className="animate-spin" aria-hidden="true" />
        running{elapsed !== null ? ` ${formatElapsed(elapsed)}` : ''}
      </span>
    );
  }
  if (status === 'error') {
    return (
      <span className="chip !border-danger/35 !text-danger">
        <CircleAlert size={11} aria-hidden="true" /> failed
      </span>
    );
  }
  if (status === 'unfinished') {
    return <span className="chip !text-ink-faint">no result</span>;
  }
  return <span className="chip !text-ink-faint">ok</span>;
}

function ToolCard({ call, defaultOpen }: { call: ToolCallView; defaultOpen: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  const Icon = toolIcon(call.name);

  return (
    <div
      className={clsx(
        'overflow-hidden rounded-xl border bg-panel-2/40',
        call.status === 'error' ? 'border-danger/30' : 'border-edge',
      )}
    >
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
        className="flex min-h-9 w-full items-center gap-2 px-3 py-2 text-left transition hover:bg-panel-2/70"
      >
        <Icon size={13} className="shrink-0 text-ink-faint" aria-hidden="true" />
        <span className="shrink-0 font-mono text-[11px] font-medium text-ink">{call.name}</span>
        {call.preview && (
          <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-ink-faint">{call.preview}</span>
        )}
        <span className="ml-auto flex shrink-0 items-center gap-2">
          <StatusPill status={call.status} since={call.at} />
          <ChevronDown
            size={13}
            className={clsx('text-ink-faint transition-transform', open && 'rotate-180')}
            aria-hidden="true"
          />
        </span>
      </button>
      {open && (
        <div className="space-y-2 border-t border-edge px-3 py-2.5">
          {call.args && (
            <div>
              <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-ink-faint">Arguments</p>
              <pre className="max-h-56 overflow-auto whitespace-pre-wrap break-words rounded-lg border border-edge bg-base p-2.5 font-mono text-[11px] leading-relaxed text-ink-dim">
                {call.args}
              </pre>
            </div>
          )}
          {call.result !== undefined && (
            <div>
              <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-ink-faint">
                {call.status === 'error' ? 'Error' : 'Result'}
              </p>
              <pre
                className={clsx(
                  'max-h-80 overflow-auto whitespace-pre-wrap break-words rounded-lg border p-2.5 font-mono text-[11px] leading-relaxed',
                  call.status === 'error'
                    ? 'border-danger/25 bg-danger/[0.05] text-danger'
                    : 'border-edge bg-base text-ink-dim',
                )}
              >
                {call.result || '(empty)'}
              </pre>
            </div>
          )}
          {call.result === undefined && call.status !== 'running' && (
            <p className="text-[11px] text-ink-faint">
              pi never wrote a result for this call — the run was interrupted or the file was deleted.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function ThinkingBlock({ text }: { text: string }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-xl border border-edge/70 bg-panel-2/25">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
        className="flex min-h-8 w-full items-center gap-2 px-3 py-1.5 text-left transition hover:bg-panel-2/60"
      >
        <Brain size={12} className="shrink-0 text-ink-faint" aria-hidden="true" />
        <span className="text-[11px] font-medium text-ink-dim">Thinking</span>
        <span className="text-[10px] tabular-nums text-ink-faint">{text.length.toLocaleString()} chars</span>
        <ChevronDown
          size={12}
          className={clsx('ml-auto text-ink-faint transition-transform', open && 'rotate-180')}
          aria-hidden="true"
        />
      </button>
      {open && (
        <p className="border-t border-edge/70 px-3 py-2.5 text-xs leading-5 text-ink-faint italic whitespace-pre-wrap break-words">
          {text}
        </p>
      )}
    </div>
  );
}

function RawBlock({ type, raw }: { type: string; raw: unknown }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-xl border border-edge/60">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
        className="flex min-h-8 w-full items-center gap-2 px-3 py-1.5 text-left transition hover:bg-panel-2/60"
      >
        <ScrollText size={12} className="shrink-0 text-ink-faint" aria-hidden="true" />
        <span className="font-mono text-[11px] text-ink-dim">{type}</span>
        <ChevronDown
          size={12}
          className={clsx('ml-auto text-ink-faint transition-transform', open && 'rotate-180')}
          aria-hidden="true"
        />
      </button>
      {open && (
        <pre className="max-h-64 overflow-auto border-t border-edge/60 px-3 py-2.5 font-mono text-[11px] text-ink-faint">
          {indentJson(raw)}
        </pre>
      )}
    </div>
  );
}

function MetaLine({ text, detail, at }: { text: string; detail?: string; at?: string }) {
  return (
    <div className="flex items-center gap-2 py-1 text-[11px] text-ink-faint">
      <span className="h-px flex-1 bg-edge/70" aria-hidden="true" />
      <span className="shrink-0">
        {text}
        {detail ? <span className="ml-1.5 font-mono text-ink-dim">{detail}</span> : null}
      </span>
      {at && (
        <time className="shrink-0 tabular-nums" dateTime={at}>
          {formatTime(at)}
        </time>
      )}
      <span className="h-px flex-1 bg-edge/70" aria-hidden="true" />
    </div>
  );
}

// ─── transcript ──────────────────────────────────────────────────────

export function SessionTranscript({
  entries,
  live,
  hasMoreBefore,
  loadingEarlier,
  onLoadEarlier,
  emptyHint,
}: {
  entries: SessionEntry[];
  live: boolean;
  hasMoreBefore: boolean;
  loadingEarlier: boolean;
  onLoadEarlier: () => void;
  emptyHint?: string;
}) {
  const blocks = useMemo(() => buildBlocks(entries, live), [entries, live]);
  const scrollerRef = useRef<HTMLDivElement>(null);
  const [following, setFollowing] = useState(true);

  useEffect(() => {
    if (!following) return;
    const el = scrollerRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [blocks.length, following]);

  const handleScroll = () => {
    const el = scrollerRef.current;
    if (!el) return;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 96;
    setFollowing(atBottom);
  };

  const jumpToLatest = () => {
    setFollowing(true);
    const el = scrollerRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  };

  return (
    <div className="relative">
      <div
        ref={scrollerRef}
        onScroll={handleScroll}
        tabIndex={0}
        role="log"
        aria-label="Agent transcript"
        aria-live="polite"
        className="h-[calc(100vh-21rem)] min-h-[420px] overflow-y-auto overscroll-contain px-4 py-4 sm:px-5"
      >
        <div className="mx-auto flex max-w-3xl flex-col gap-4">
          {hasMoreBefore && (
            <div className="flex justify-center">
              <button type="button" className="btn btn-ghost" onClick={onLoadEarlier} disabled={loadingEarlier}>
                {loadingEarlier ? <Loader2 size={13} className="animate-spin" /> : <ArrowDownToLine size={13} />}
                {loadingEarlier ? 'Loading earlier…' : 'Load earlier messages'}
              </button>
            </div>
          )}

          {blocks.length === 0 && (
            <p className="px-2 py-10 text-center text-xs leading-5 text-ink-faint">
              {emptyHint ?? 'No messages yet.'}
            </p>
          )}

          {blocks.map((block) => {
            switch (block.kind) {
              case 'user':
                return (
                  <div key={block.key} className="flex justify-end gap-2.5">
                    <div className="max-w-[85%] rounded-2xl border border-accent/25 bg-accent/[0.07] px-3.5 py-2.5">
                      <p className="mb-1 flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-accent">
                        <User size={10} aria-hidden="true" /> Prompt
                      </p>
                      <p className="text-sm leading-6 whitespace-pre-wrap break-words text-ink">{block.text}</p>
                    </div>
                  </div>
                );
              case 'assistant':
                return (
                  <div key={block.key} className="flex gap-2.5">
                    <span
                      className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-edge bg-panel-2 text-ink-faint"
                      aria-hidden="true"
                    >
                      <Bot size={13} />
                    </span>
                    <div className="min-w-0 flex-1 space-y-2">
                      {block.parts.map((part, index) => {
                        if (part.kind === 'text') {
                          return (
                            <p
                              key={`${block.key}:t${index}`}
                              className="text-sm leading-6 whitespace-pre-wrap break-words text-ink"
                            >
                              {part.text}
                            </p>
                          );
                        }
                        if (part.kind === 'thinking') {
                          return <ThinkingBlock key={`${block.key}:k${index}`} text={part.text} />;
                        }
                        return (
                          <ToolCard
                            key={`${block.key}:c${part.call.id}`}
                            call={part.call}
                            defaultOpen={part.call.status === 'error' || part.call.status === 'running'}
                          />
                        );
                      })}
                      {block.error && (
                        <div className="rounded-lg border border-danger/30 bg-danger/[0.06] px-3 py-2 text-xs text-danger">
                          {block.error}
                        </div>
                      )}
                      {block.stopReason === 'error' && !block.error && (
                        <div className="rounded-lg border border-danger/30 bg-danger/[0.06] px-3 py-2 text-xs text-danger">
                          The model ended this turn with an error.
                        </div>
                      )}
                    </div>
                  </div>
                );
              case 'bash':
                return (
                  <div key={block.key} className="rounded-xl border border-edge bg-panel-2/40">
                    <p className="flex items-center gap-2 border-b border-edge px-3 py-2 font-mono text-[11px] text-ink-dim">
                      <Terminal size={12} aria-hidden="true" /> {block.command}
                      {block.exitCode !== undefined && block.exitCode !== 0 && (
                        <span className="text-danger">exit {block.exitCode}</span>
                      )}
                    </p>
                    <pre className="max-h-72 overflow-auto px-3 py-2.5 font-mono text-[11px] leading-relaxed whitespace-pre-wrap break-words text-ink-faint">
                      {block.output || '(no output)'}
                    </pre>
                  </div>
                );
              case 'system':
                return <RawBlock key={block.key} type="prompt context" raw={block.text} />;
              case 'compaction':
                return (
                  <div key={block.key} className="rounded-xl border border-violet/25 bg-violet/[0.05] px-3.5 py-3">
                    <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-violet">
                      Context compacted
                    </p>
                    <p className="text-xs leading-5 whitespace-pre-wrap break-words text-ink-dim">{block.text}</p>
                  </div>
                );
              case 'meta':
                return <MetaLine key={block.key} text={block.text} detail={block.detail} at={block.at} />;
              case 'raw':
                return <RawBlock key={block.key} type={block.type} raw={block.raw} />;
              default:
                return null;
            }
          })}
        </div>
      </div>

      {!following && (
        <div className="pointer-events-none absolute inset-x-0 bottom-3 flex justify-center">
          <button
            type="button"
            onClick={jumpToLatest}
            className="pointer-events-auto btn glass-chrome glass-control shadow-lg"
          >
            <ArrowDownToLine size={14} /> Jump to latest
          </button>
        </div>
      )}
    </div>
  );
}
