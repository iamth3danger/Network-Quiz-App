"use client";

import { useParams } from "next/navigation";
import QuizRunner from "@/components/QuizRunner";

export default function QuizPage() {
  const { objectiveid } = useParams<{ objectiveid: string }>();

  return (
    <QuizRunner
      title="Quiz"
      query={`exam_objective=${encodeURIComponent(objectiveid)}`}
    />
  );
}
