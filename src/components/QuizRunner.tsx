"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Link from "next/link";
import { apiUrl } from "@/lib/api";

interface ApiError {
  status: number;
  error: string;
}

interface ApiResponse<T> {
  status: number;
  data: T;
}

interface Question {
  id: number;
  objective_id: number;
  subject_id: number;
  question: string;
  a: string;
  b: string;
  c: string;
  d: string;
  answer: string;
}

type ApiResult<T> = ApiResponse<T> | ApiError;

// `query` is the backend selector for this quiz, either
// "exam_objective=1.1" or "subject=tcp_flags".
export default function QuizRunner({
  title,
  query,
}: {
  title: string;
  query: string;
}) {
  const [selected, setSelected] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);
  const [question, setQuestion] = useState<Question | null>(null);
  const [questionNumber, setQuestionNumber] = useState<number | null>(null);

  const fetchQuestionNumber = useCallback(async () => {
    try {
      const res = await fetch(apiUrl(`/quiz/count?${query}`));
      const data = (await res.json()) as ApiResult<{ Count: number }>;
      if ("data" in data) {
        setQuestionNumber(data.data.Count);
      }
    } catch (err) {
      console.error(err);
    }
  }, [query]);

  const fetchQuestion = useCallback(async () => {
    setQuestion(null);
    setSelected(null);
    setSubmitted(false);
    await fetchQuestionNumber();
    try {
      let res = await fetch(apiUrl(`/questions?${query}`));
      let data = (await res.json()) as ApiResult<Question>;
      if ("error" in data && data.status === 404) {
        res = await fetch(apiUrl(`/questions/generate?${query}`), {
          method: "POST",
        });
        data = (await res.json()) as ApiResult<Question>;
      }
      if ("error" in data) {
        throw new Error((data as ApiError).error);
      }
      setQuestion(data.data);
    } catch (err) {
      console.error(err);
    }
  }, [query, fetchQuestionNumber]);

  // Fetch once per quiz — the ref keeps StrictMode's double-effect from
  // burning a question, but still refetches when the route switches quizzes.
  const fetchedFor = useRef<string | null>(null);
  useEffect(() => {
    if (fetchedFor.current === query) return;
    fetchedFor.current = query;
    fetchQuestion();
  }, [query, fetchQuestion]);

  const handleSubmit = async (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selected || !question) return;

    await fetch(apiUrl("/questions"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        question: question.question,
        question_id: question.id,
        correct: selected === question.answer,
        exam_objective_id: question.objective_id,
        subject_id: question.subject_id,
      }),
    });

    setSubmitted(true);
  };

  const isCorrect = selected === question?.answer;

  const options: [string, string][] = question
    ? [
        ["a", question.a],
        ["b", question.b],
        ["c", question.c],
        ["d", question.d],
      ]
    : [];

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900 p-4">
      <div className="w-full max-w-lg bg-gray-800 rounded-2xl shadow-xl p-8">
        <div className="flex justify-between items-center mb-6">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="text-gray-400 hover:text-white transition-colors"
            >
              &larr;
            </Link>
            <h1 className="text-2xl font-bold text-white">{title}</h1>
          </div>
          {questionNumber !== null && (
            <span className="text-gray-400 text-sm font-medium">
              Question {questionNumber} of 10
            </span>
          )}
        </div>
        {question ? (
          <form onSubmit={handleSubmit}>
            <p className="text-lg text-gray-100 mb-6">{question.question}</p>
            <div className="space-y-3 mb-8">
              {options.map(([key, value]) => (
                <label
                  key={key}
                  className={`flex items-center gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                    selected === key
                      ? "border-blue-500 bg-blue-500/10"
                      : "border-gray-600 hover:border-gray-400"
                  } ${
                    submitted && key === question.answer
                      ? "border-green-500 bg-green-500/10"
                      : ""
                  } ${
                    submitted && selected === key && !isCorrect
                      ? "border-red-500 bg-red-500/10"
                      : ""
                  }`}
                >
                  <input
                    type="radio"
                    name="answer"
                    value={key}
                    checked={selected === key}
                    onChange={() => setSelected(key)}
                    disabled={submitted}
                    className="accent-blue-500"
                  />
                  <span className="text-gray-200 font-medium uppercase">
                    {key}.
                  </span>
                  <span className="text-gray-200">{value}</span>
                </label>
              ))}
            </div>
            {!submitted ? (
              <button
                type="submit"
                disabled={!selected}
                className="w-full py-3 rounded-lg bg-blue-600 text-white font-semibold hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                Submit Answer
              </button>
            ) : (
              <div className="space-y-4">
                {isCorrect ? (
                  <div className="p-4 rounded-lg bg-green-500/20 border border-green-500 text-green-300 text-center font-semibold">
                    Correct! Nice work.
                  </div>
                ) : (
                  <div className="p-4 rounded-lg bg-red-500/20 border border-red-500 text-red-300 text-center">
                    <p className="font-semibold">Wrong answer!</p>
                    <p className="mt-1 text-sm">
                      The correct answer was{" "}
                      <span className="font-bold uppercase">
                        {question.answer}
                      </span>
                      : {question[question.answer as keyof Question]}
                    </p>
                  </div>
                )}
                <button
                  type="button"
                  onClick={fetchQuestion}
                  className="w-full py-3 rounded-lg bg-gray-600 text-white font-semibold hover:bg-gray-500 transition-colors"
                >
                  Next Question
                </button>
              </div>
            )}
          </form>
        ) : (
          <p className="text-gray-400">Loading question...</p>
        )}
      </div>
    </div>
  );
}
