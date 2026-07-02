"use client";

import { useState, useEffect, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";

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
  question: string;
  a: string;
  b: string;
  c: string;
  d: string;
  answer: string;
}

type ApiResult<T> = ApiResponse<T> | ApiError;

export default function QuizPage() {
  const { objectiveid } = useParams<{ objectiveid: string }>();
  const [selected, setSelected] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);
  const [question, setQuestion] = useState<Question | null>(null);
  const [questionNumber, setQuestionNumber] = useState<number | null>(null);

  const fetchQuestionNumber = async () => {
    try {
      const res = await fetch(
        `http://localhost:8080/quiz/count?exam_objective=${objectiveid}`,
      );
      const data = (await res.json()) as ApiResult<{ Count: number }>;
      if ("data" in data) {
        setQuestionNumber(data.data.Count);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const fetchQuestion = async () => {
    setQuestion(null);
    setSelected(null);
    setSubmitted(false);
    await fetchQuestionNumber();
    try {
      let res = await fetch(
        `http://localhost:8080/questions?exam_objective=${objectiveid}`,
      );
      let data = (await res.json()) as ApiResult<Question>;
      if ("error" in data && data.status === 404) {
        res = await fetch(
          `http://localhost:8080/questions/generate?exam_objective=${objectiveid}`,
          { method: "POST" },
        );
        data = (await res.json()) as ApiResult<Question>;
      }
      if ("error" in data) {
        throw new Error((data as ApiError).error);
      }
      setQuestion(data.data);
    } catch (err) {
      console.error(err);
    }
  };

  const didFetch = useRef(false);
  useEffect(() => {
    if (didFetch.current) return;
    didFetch.current = true;
    fetchQuestion();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected || !question) return;

    await fetch("http://localhost:8080/questions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        question_id: question.id,
        correct: selected === question.answer,
        exam_objective_id: question.objective_id,
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
            <h1 className="text-2xl font-bold text-white">Quiz</h1>
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
