import { useAuth } from "../../modules/auth/model/useAuth";

export default function SubscriptionRequiredPage() {
  const { signOut } = useAuth();

  return (
    <main className="screen">
      <section className="welcome-card">
        <h1>Подписка неактивна</h1>
        <p>
          Для доступа к рабочим экранам нужна активная платная подписка.
          Обратитесь к администратору.
        </p>
        <button className="logout-button" type="button" onClick={signOut}>
          Выйти
        </button>
      </section>
    </main>
  );
}
