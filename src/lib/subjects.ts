// Subject quizzes, one per row in the backend `subjects` table. `slug` is what
// goes on the wire (the title with spaces swapped for underscores); the backend
// swaps them back before looking the title up (see ser/cmd/main.go parseSubject).
export interface Subject {
  slug: string;
  title: string;
  label: string;
}

const titles: [title: string, label: string][] = [
  ["subnet conversion", "Subnet Conversion"],
  ["ethernet standards", "Ethernet Standards"],
  ["wifi standards", "Wi-Fi Standards"],
  ["dhcp features", "DHCP Features"],
  ["antennas", "Antennas"],
  ["auth protocol", "Auth Protocols"],
  ["dns records", "DNS Records"],
  ["switch errors", "Switch Errors"],
  ["tcp flags", "TCP Flags"],
];

export const subjects: Subject[] = titles.map(([title, label]) => ({
  slug: title.replaceAll(" ", "_"),
  title,
  label,
}));

export function subjectBySlug(slug: string): Subject | undefined {
  return subjects.find((s) => s.slug === slug);
}

// Query-string key the backend reads the subject title from.
export const SUBJECT_PARAM = "subject";

// e.g. subjectQuery("tcp_flags") -> "subject=tcp_flags"
export function subjectQuery(slug: string): string {
  return `${SUBJECT_PARAM}=${encodeURIComponent(slug)}`;
}
