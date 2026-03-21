import { ArrowRight, LoaderCircle, LockKeyhole, Mail } from "lucide-react";
import { useState } from "react";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validateForm({ identifier, password }) {
  const normalizedIdentifier = identifier.trim();

  if (!normalizedIdentifier) return "Email или логин обязательны";
  if (normalizedIdentifier.includes("@") && !EMAIL_PATTERN.test(normalizedIdentifier)) {
    return "Введите корректный email";
  }
  if (!normalizedIdentifier.includes("@") && normalizedIdentifier.length < 3) {
    return "Введите логин администратора";
  }

  if (!password.trim()) return "Пароль обязателен";

  return "";
}

export default function LoginForm({ onSignIn, isSigningIn, signInError }) {
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [localError, setLocalError] = useState("");
  const loading = isSigningIn;
  const error = signInError;

  async function handleSubmit(event) {
    event.preventDefault();

    const validationError = validateForm({
      identifier,
      password,
    });
    setLocalError(validationError);

    if (validationError) {
      return;
    }

    await onSignIn({ identifier: identifier.trim(), password });
  }

  return (
    <section className="auth">
      <div className="auth-shell">
        <div className="auth-panel auth-panel--form">
          <div className="auth-brand">
            <span className="auth-brand__mark" aria-hidden="true" />
            <span className="auth-brand__text">TitanIT ProfDNK</span>
          </div>

          <div className="auth-copy">
            <h1 className="auth-copy__title">С возвращением</h1>
            <p className="auth-copy__subtitle">
              Психологи входят по рабочему email, администратор по логину.
              Пароль задаётся внутри платформы.
            </p>
          </div>

          <form className="auth-form" onSubmit={handleSubmit}>
            <div className="auth-form__field">
              <label className="auth-form__label" htmlFor="identifier">
                Email / логин
              </label>
              <div className="auth-form__control">
                <Mail size={16} strokeWidth={2.1} />
                <input
                  className="auth-form__input"
                  id="identifier"
                  type="text"
                  value={identifier}
                  onChange={(event) => {
                    setIdentifier(event.target.value);
                    if (localError) setLocalError("");
                  }}
                  autoComplete="username"
                  placeholder="name@example.com или admin123"
                />
              </div>
            </div>

            <div className="auth-form__field">
              <label className="auth-form__label" htmlFor="password">
                Пароль
              </label>
              <div className="auth-form__control">
                <LockKeyhole size={16} strokeWidth={2.1} />
                <input
                  className="auth-form__input"
                  id="password"
                  type="password"
                  value={password}
                  onChange={(event) => {
                    setPassword(event.target.value);
                    if (localError) setLocalError("");
                  }}
                  autoComplete="current-password"
                  placeholder="Введите пароль"
                />
              </div>
            </div>

            {localError ? <p className="auth__error">{localError}</p> : null}
            {error?.message ? (
              <p className="auth__error">{error.message}</p>
            ) : null}

            <button
              className="auth-form__submit"
              type="submit"
              disabled={loading}
            >
              <span>{loading ? "Проверяем доступ" : "Войти"}</span>
              <span className="auth-form__submit-icon" aria-hidden="true">
                {loading ? (
                  <LoaderCircle
                    className="icon-spin"
                    size={16}
                    strokeWidth={2.2}
                  />
                ) : (
                  <ArrowRight size={16} strokeWidth={2.4} />
                )}
              </span>
            </button>
          </form>
        </div>
        <div className="auth-panel auth-panel--visual">
          <div className="auth-visual">
            <div className="auth-visual__flare" />
            <p className="auth-visual__eyebrow">ProfDNK platform</p>
            <h2 className="auth-visual__title">
              Платформа
              <br />
              <span className="auth-visual__title-line">профориентации и</span>
              <br />
              психодиагностики
            </h2>
            <p className="auth-visual__text">
              Создание методик, клиентские сессии, результаты и отчёты в одном
              рабочем контуре.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
