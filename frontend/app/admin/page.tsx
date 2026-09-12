import { redirect } from "next/navigation";

/**
 * /admin — lands on the first live ops section.
 * Add more sections to the sidebar in admin/layout.tsx as they ship.
 */
export default function AdminHomePage() {
  redirect("/admin/ops");
}
