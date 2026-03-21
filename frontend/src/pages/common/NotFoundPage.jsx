import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

export default function NotFoundPage() {
  return (
    <PageCard
      title="404"
      description="Страница не найдена."
      links={[{ to: ROUTES.root, label: "На главную" }]}
    />
  );
}
