import { ROUTES } from "../../shared/config/routes";
import PageCard from "../../shared/ui/PageCard";

export default function ForbiddenPage() {
  return (
    <PageCard
      title="Доступ запрещен"
      description="У вас нет прав для просмотра этой страницы."
      links={[{ to: ROUTES.root, label: "На главную" }]}
    />
  );
}
