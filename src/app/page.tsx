"use client";

import { useState, useEffect } from "react";
import Link from "next/link";

interface Objective {
  code: string;
  description: string;
}

interface ObjectiveRating {
  code: string;
  stars: number;
}

interface TermStat {
  code: string;
  term_name: string;
  right_count: number;
  wrong_count: number;
}

const objectives: Objective[] = [
  {
    code: "1.1",
    description:
      "Explain concepts related to the Open Systems Interconnection (OSI) reference model.",
  },
  {
    code: "1.2",
    description:
      "Compare and contrast networking appliances, applications, and functions.",
  },
  {
    code: "1.3",
    description: "Summarize cloud concepts and connectivity options.",
  },
  {
    code: "1.4",
    description:
      "Explain common networking ports, protocols, services, and traffic types.",
  },
  {
    code: "1.5",
    description: "Compare and contrast transmission media and transceivers.",
  },
  {
    code: "1.6",
    description:
      "Compare and contrast network topologies, architectures, and types.",
  },
  {
    code: "1.7",
    description: "Given a scenario, use appropriate IPv4 network addressing.",
  },
  {
    code: "1.8",
    description:
      "Summarize evolving use cases for modern network environments.",
  },
  {
    code: "2.1",
    description: "Explain characteristics of routing technologies.",
  },
  {
    code: "2.2",
    description:
      "Given a scenario, configure switching technologies and features.",
  },
  {
    code: "2.3",
    description:
      "Given a scenario, select and configure wireless devices and technologies.",
  },
  {
    code: "2.4",
    description: "Explain important factors of physical installations.",
  },
  {
    code: "3.1",
    description:
      "Explain the purpose of organizational processes and procedures.",
  },
  {
    code: "3.2",
    description: "Given a scenario, use network monitoring technologies.",
  },
  { code: "3.3", description: "Explain disaster recovery (DR) concepts." },
  {
    code: "3.4",
    description: "Given a scenario, implement IPv4 and IPv6 network services.",
  },
  {
    code: "3.5",
    description: "Compare and contrast network access and management methods.",
  },
  {
    code: "4.1",
    description: "Explain the importance of basic network security concepts.",
  },
  {
    code: "4.2",
    description:
      "Summarize various types of attacks and their impact to the network.",
  },
  {
    code: "4.3",
    description:
      "Given a scenario, apply network security features, defense techniques, and solutions.",
  },
  { code: "5.1", description: "Explain the troubleshooting methodology." },
  {
    code: "5.2",
    description:
      "Given a scenario, troubleshoot common cabling and physical interface issues.",
  },
  {
    code: "5.3",
    description:
      "Given a scenario, troubleshoot common issues with network services.",
  },
  {
    code: "5.4",
    description: "Given a scenario, troubleshoot common performance issues.",
  },
  {
    code: "5.5",
    description:
      "Given a scenario, use the appropriate tool or protocol to solve networking issues.",
  },
];

export default function Home() {
  const [expandedDomain, setExpandedDomain] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [termStats, setTermStats] = useState<TermStat[]>([]);
  const [ratings, setRatings] = useState<ObjectiveRating[]>([]);

  useEffect(() => {
    fetch("http://localhost:8080/terms")
      .then((res) => res.json())
      .then((data) => {
        if (data.data) setTermStats(data.data);
      })
      .catch(console.error);

    fetch("http://localhost:8080/rating")
      .then((res) => res.json())
      .then((data) => {
        if (data.data) setRatings(data.data);
      })
      .catch(console.error);
  }, []);

  const toggleDomain = (label: string) => {
    setExpandedDomain(expandedDomain === label ? null : label);
  };

  const toggle = (code: string) => {
    setExpanded(expanded === code ? null : code);
  };

  const domains = [
    { label: "1.0 Networking Concepts", range: ["1.1", "1.8"] },
    { label: "2.0 Network Implementation", range: ["2.1", "2.4"] },
    { label: "3.0 Network Operations", range: ["3.1", "3.5"] },
    { label: "4.0 Network Security", range: ["4.1", "4.3"] },
    { label: "5.0 Network Troubleshooting", range: ["5.1", "5.5"] },
  ];

  return (
    <div className="min-h-screen bg-gray-900 p-4 md:p-8">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold text-white mb-8">
          CompTIA Network+ N10-009
        </h1>

        <Link
          href="/matching"
          className="block mb-6 p-4 bg-gray-800 rounded-lg border border-gray-700 hover:border-blue-500 transition-colors"
        >
          <span className="text-blue-400 font-semibold">Ports</span>
          <span className="text-gray-400 ml-2 text-sm">
            Match port numbers to protocols
          </span>
        </Link>

        {domains.map((domain) => {
          const domainObjectives = objectives.filter(
            (o) => o.code >= domain.range[0] && o.code <= domain.range[1],
          );

          const isDomainOpen = expandedDomain === domain.label;

          return (
            <div key={domain.label} className="mb-4">
              <button
                onClick={() => toggleDomain(domain.label)}
                className="w-full flex items-center gap-2 text-left text-lg font-semibold text-gray-300 border-b border-gray-700 pb-2 mb-3 hover:text-white transition-colors"
              >
                <span
                  className={`text-gray-400 transition-transform text-sm ${
                    isDomainOpen ? "rotate-90" : ""
                  }`}
                >
                  &#9654;
                </span>
                {domain.label}
              </button>
              {isDomainOpen && (
                <div className="space-y-2">
                  {domainObjectives.map((obj) => {
                    const terms = termStats.filter((t) => t.code === obj.code);
                    const totalRight = terms.reduce((s, t) => s + Number(t.right_count), 0);
                    const totalWrong = terms.reduce((s, t) => s + Number(t.wrong_count), 0);
                    const isOpen = expanded === obj.code;
                    const rating = ratings.find((r) => r.code === obj.code);
                    const starCount = rating ? rating.stars : 0;

                    return (
                      <div
                        key={obj.code}
                        className="bg-gray-800 rounded-lg border border-gray-700"
                      >
                        <div className="flex items-center justify-between p-4">
                          <button
                            onClick={() => toggle(obj.code)}
                            className="flex items-center gap-2 text-left flex-1 min-w-0"
                          >
                            <span
                              className={`text-gray-400 transition-transform text-sm ${
                                isOpen ? "rotate-90" : ""
                              }`}
                            >
                              &#9654;
                            </span>
                            <span className="text-gray-400 font-mono text-sm shrink-0">
                              {obj.code}
                            </span>
                            {starCount > 0 && (
                              <span className="text-yellow-400 text-sm shrink-0">
                                {"★".repeat(starCount)}
                              </span>
                            )}
                            <Link
                              href={`/quiz/${obj.code}`}
                              onClick={(e) => e.stopPropagation()}
                              className="text-blue-400 hover:text-blue-300 hover:underline truncate"
                            >
                              {obj.description}
                            </Link>
                          </button>
                          <div className="flex items-center ml-4 shrink-0 rounded-full overflow-hidden text-xs font-medium">
                            <span className="bg-green-500/20 text-green-400 px-2 py-0.5">
                              {totalRight}
                            </span>
                            <span className="bg-red-500/20 text-red-400 px-2 py-0.5">
                              {totalWrong}
                            </span>
                          </div>
                        </div>

                        {isOpen && terms.length > 0 && (
                          <div className="border-t border-gray-700 px-4 pb-3 pt-2">
                            <div className="space-y-1">
                              {terms
                                .sort((a, b) => (Number(b.right_count) + Number(b.wrong_count)) - (Number(a.right_count) + Number(a.wrong_count)))
                                .map((term) => (
                                <div
                                  key={term.term_name}
                                  className="flex items-center justify-between py-1.5 px-3 rounded hover:bg-gray-700/50"
                                >
                                  <span className="text-gray-300 text-sm">
                                    {term.term_name}
                                  </span>
                                  <div className="flex items-center rounded-full overflow-hidden text-xs font-medium">
                                    <span className="bg-green-500/15 text-green-400/70 px-1.5 py-0.5">
                                      {Number(term.right_count)}
                                    </span>
                                    <span className="bg-red-500/15 text-red-400/70 px-1.5 py-0.5">
                                      {Number(term.wrong_count)}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
