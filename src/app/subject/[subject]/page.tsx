"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import QuizRunner from "@/components/QuizRunner";
import { subjectBySlug, subjectQuery } from "@/lib/subjects";

export default function SubjectQuizPage() {
  const { subject: slug } = useParams<{ subject: string }>();
  const subject = subjectBySlug(slug);

  if (!subject) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 p-4">
        <div className="w-full max-w-lg bg-gray-800 rounded-2xl shadow-xl p-8 text-center">
          <p className="text-gray-200 mb-4">Unknown subject &quot;{slug}&quot;.</p>
          <Link href="/" className="text-blue-400 hover:text-blue-300">
            Back to subjects
          </Link>
        </div>
      </div>
    );
  }

  return (
    <QuizRunner title={subject.label} query={subjectQuery(subject.slug)} />
  );
}
