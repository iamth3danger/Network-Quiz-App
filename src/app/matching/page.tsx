"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { apiUrl } from "@/lib/api";

interface MatchingRow {
  id: number;
  abbreviation: string;
  port: string;
}

interface MatchResult {
  port_id: number;
  correct: boolean;
}

function shuffle<T>(arr: T[]): T[] {
  const a = [...arr];
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

export default function MatchingPage() {
  const [rows, setRows] = useState<MatchingRow[]>([]);
  const [shuffledPorts, setShuffledPorts] = useState<MatchingRow[]>([]);
  const [selectedAbbr, setSelectedAbbr] = useState<number | null>(null);
  const [selectedPort, setSelectedPort] = useState<number | null>(null);
  const [matched, setMatched] = useState<Set<number>>(new Set());
  const [results, setResults] = useState<MatchResult[]>([]);
  const [wrong, setWrong] = useState<{ abbr: number; port: number } | null>(null);
  const [wrongRecorded, setWrongRecorded] = useState<Set<number>>(new Set());
  const [done, setDone] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchRows = useCallback(async () => {
    setLoading(true);
    setMatched(new Set());
    setResults([]);
    setSelectedAbbr(null);
    setSelectedPort(null);
    setWrong(null);
    setWrongRecorded(new Set());
    setDone(false);
    try {
      const res = await fetch(apiUrl("/matching/generate"));
      const data = await res.json();
      const items: MatchingRow[] = data.data ?? data;
      setRows(shuffle(items));
      setShuffledPorts(shuffle(items));
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRows();
  }, [fetchRows]);

  const tryMatch = useCallback(
    (abbrId: number | null, portId: number | null) => {
      if (abbrId === null || portId === null) return;

      const isCorrect = abbrId === portId;

      if (isCorrect) {
        const result: MatchResult = { port_id: abbrId, correct: !wrongRecorded.has(abbrId) };
        const newResults = [...results, result];
        const newMatched = new Set(matched);
        newMatched.add(abbrId);
        setMatched(newMatched);
        setSelectedAbbr(null);
        setSelectedPort(null);

        if (newMatched.size === rows.length) {
          setResults(newResults);
          setDone(true);
          fetch(apiUrl("/matching"), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(newResults),
          }).catch(console.error);
        } else {
          setResults(newResults);
        }
      } else {
        if (!wrongRecorded.has(abbrId)) {
          const newWrongRecorded = new Set(wrongRecorded);
          newWrongRecorded.add(abbrId);
          setWrongRecorded(newWrongRecorded);
        }
        setWrong({ abbr: abbrId, port: portId });
        setTimeout(() => {
          setWrong(null);
          setSelectedAbbr(null);
          setSelectedPort(null);
        }, 800);
      }
    },
    [matched, results, rows.length, wrongRecorded],
  );

  const handleAbbrClick = (id: number) => {
    if (matched.has(id) || (wrong && wrong.abbr === id)) return;
    setSelectedAbbr(id);
    tryMatch(id, selectedPort);
  };

  const handlePortClick = (id: number) => {
    if (matched.has(id) || (wrong && wrong.port === id)) return;
    setSelectedPort(id);
    tryMatch(selectedAbbr, id);
  };

  const correctCount = results.filter((r) => r.correct).length;
  const wrongCount = done
    ? results.filter((r) => !r.correct).length
    : wrongRecorded.size;

  return (
    <div className="min-h-screen bg-gray-900 p-4 md:p-8">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="text-gray-400 hover:text-white transition-colors"
            >
              &larr;
            </Link>
            <h1 className="text-2xl font-bold text-white">
              Ports
            </h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center rounded-full overflow-hidden text-sm font-medium">
              <span className="bg-green-500/20 text-green-400 px-2 py-0.5">
                {correctCount}
              </span>
              <span className="bg-red-500/20 text-red-400 px-2 py-0.5">
                {wrongCount}
              </span>
            </div>
          </div>
        </div>

        {loading ? (
          <p className="text-gray-400">Loading...</p>
        ) : done ? (
          <div className="space-y-6">
            <div className="p-6 rounded-lg bg-green-500/20 border border-green-500 text-green-300 text-center">
              <p className="text-xl font-semibold">All matched!</p>
              <p className="mt-2 text-sm">
                {correctCount} correct, {wrongCount} wrong attempts
              </p>
            </div>
            <button
              onClick={fetchRows}
              className="w-full py-3 rounded-lg bg-blue-600 text-white font-semibold hover:bg-blue-700 transition-colors"
            >
              Play Again
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-6">
            <div>
              <h2 className="text-sm font-medium text-gray-400 mb-3 uppercase tracking-wide">
                Protocol
              </h2>
              <div className="space-y-2">
                {rows.map((row) => {
                  const isMatched = matched.has(row.id);
                  const isSelected = selectedAbbr === row.id;
                  const isWrong = wrong?.abbr === row.id;

                  return (
                    <button
                      key={row.id}
                      onClick={() => handleAbbrClick(row.id)}
                      disabled={isMatched || wrong !== null}
                      className={`w-full text-left p-3 rounded-lg border transition-colors ${
                        isMatched
                          ? "border-green-500/30 bg-green-500/10 text-green-400/50"
                          : isWrong
                            ? "border-red-500 bg-red-500/10 text-red-300"
                            : isSelected
                              ? "border-blue-500 bg-blue-500/10 text-white"
                              : "border-gray-700 hover:border-gray-500 text-gray-200"
                      }`}
                    >
                      {row.abbreviation}
                    </button>
                  );
                })}
              </div>
            </div>

            <div>
              <h2 className="text-sm font-medium text-gray-400 mb-3 uppercase tracking-wide">
                Port
              </h2>
              <div className="space-y-2">
                {shuffledPorts.map((row) => {
                  const isMatched = matched.has(row.id);
                  const isSelected = selectedPort === row.id;
                  const isWrong = wrong?.port === row.id;

                  return (
                    <button
                      key={row.id}
                      onClick={() => handlePortClick(row.id)}
                      disabled={isMatched || wrong !== null}
                      className={`w-full text-left p-3 rounded-lg border transition-colors ${
                        isMatched
                          ? "border-green-500/30 bg-green-500/10 text-green-400/50"
                          : isWrong
                            ? "border-red-500 bg-red-500/10 text-red-300"
                            : isSelected
                              ? "border-blue-500 bg-blue-500/10 text-white"
                              : "border-gray-700 hover:border-gray-500 text-gray-200"
                      }`}
                    >
                      {row.port}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
