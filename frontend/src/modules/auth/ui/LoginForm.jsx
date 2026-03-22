import { ArrowRight, LoaderCircle, LockKeyhole, Mail } from "lucide-react";
import { useState } from "react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const TEST_ACCOUNT_LOGIN = "admin123";
const TEST_ACCOUNT_PASSWORD = "pass12345";

function createStaggerTransition(reducedMotion) {
  if (reducedMotion) {
    return {
      duration: 0.01,
      staggerChildren: 0,
      delayChildren: 0,
    };
  }

  return {
    duration: 0.5,
    ease: [0.22, 1, 0.36, 1],
    staggerChildren: 0.08,
    delayChildren: 0.08,
  };
}

function createItemVariants(reducedMotion, axis = "y", distance = 18) {
  return {
    hidden: {
      opacity: 0,
      [axis]: reducedMotion ? 0 : distance,
    },
    visible: {
      opacity: 1,
      [axis]: 0,
      transition: {
        duration: reducedMotion ? 0.01 : 0.55,
        ease: [0.22, 1, 0.36, 1],
      },
    },
  };
}

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
  const reducedMotion = useReducedMotion();
  const loading = isSigningIn;
  const error = signInError;
  const formStackVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: createStaggerTransition(reducedMotion),
    },
  };
  const leftPanelVariants = {
    hidden: {
      opacity: 0,
      x: reducedMotion ? 0 : -24,
    },
    visible: {
      opacity: 1,
      x: 0,
      transition: {
        duration: reducedMotion ? 0.01 : 0.62,
        ease: [0.22, 1, 0.36, 1],
        when: "beforeChildren",
        staggerChildren: reducedMotion ? 0 : 0.08,
      },
    },
  };
  const rightPanelVariants = {
    hidden: {
      opacity: 0,
      x: reducedMotion ? 0 : 28,
      scale: reducedMotion ? 1 : 0.985,
    },
    visible: {
      opacity: 1,
      x: 0,
      scale: 1,
      transition: {
        duration: reducedMotion ? 0.01 : 0.72,
        ease: [0.22, 1, 0.36, 1],
        when: "beforeChildren",
        staggerChildren: reducedMotion ? 0 : 0.1,
        delayChildren: reducedMotion ? 0 : 0.18,
      },
    },
  };
  const riseVariants = createItemVariants(reducedMotion, "y", 18);
  const slideVariants = createItemVariants(reducedMotion, "x", 20);
  const activeError = localError || error?.message || "";

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
    <motion.section
      animate="visible"
      className="auth"
      initial="hidden"
      variants={{
        hidden: { opacity: 0 },
        visible: {
          opacity: 1,
          transition: {
            duration: reducedMotion ? 0.01 : 0.35,
            ease: [0.22, 1, 0.36, 1],
          },
        },
      }}
    >
      <motion.div
        className="auth-shell"
        variants={{
          hidden: {
            opacity: 0,
            y: reducedMotion ? 0 : 20,
            scale: reducedMotion ? 1 : 0.985,
          },
          visible: {
            opacity: 1,
            y: 0,
            scale: 1,
            transition: {
              duration: reducedMotion ? 0.01 : 0.65,
              ease: [0.22, 1, 0.36, 1],
            },
          },
        }}
      >
        <motion.div
          className="auth-panel auth-panel--form"
          variants={leftPanelVariants}
        >
          <motion.div className="auth-brand" variants={riseVariants}>
            <span className="auth-brand__mark" aria-hidden="true" />
            <span className="auth-brand__text">TitanIT ProfDNK</span>
          </motion.div>

          <motion.div className="auth-copy" variants={riseVariants}>
            <h1 className="auth-copy__title">С возвращением</h1>
            <p className="auth-copy__subtitle">
              Психологи входят по рабочему email, администратор по логину.
              Пароль задаётся внутри платформы.
            </p>
            <p className="auth-copy__demo">
              Тестовый вход: <strong>{TEST_ACCOUNT_LOGIN}</strong> / <strong>{TEST_ACCOUNT_PASSWORD}</strong>
            </p>
          </motion.div>

          <motion.form
            className="auth-form"
            onSubmit={handleSubmit}
            variants={formStackVariants}
          >
            <motion.div className="auth-form__field" variants={riseVariants}>
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
            </motion.div>

            <motion.div className="auth-form__field" variants={riseVariants}>
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
            </motion.div>

            <AnimatePresence mode="wait">
              {activeError ? (
                <motion.p
                  animate={{ opacity: 1, y: 0 }}
                  className="auth__error"
                  exit={{ opacity: 0, y: reducedMotion ? 0 : -8 }}
                  initial={{ opacity: 0, y: reducedMotion ? 0 : -8 }}
                  key={activeError}
                  transition={{
                    duration: reducedMotion ? 0.01 : 0.22,
                    ease: [0.22, 1, 0.36, 1],
                  }}
                >
                  {activeError}
                </motion.p>
              ) : null}
            </AnimatePresence>

            <motion.button
              className="auth-form__submit"
              type="submit"
              disabled={loading}
              variants={riseVariants}
              whileHover={reducedMotion ? undefined : { y: -1, scale: 1.01 }}
              whileTap={reducedMotion ? undefined : { scale: 0.992 }}
            >
              <span>{loading ? "Проверяем доступ" : "Войти"}</span>
              <motion.span
                className="auth-form__submit-icon"
                aria-hidden="true"
                animate={
                  loading && !reducedMotion
                    ? { rotate: 360 }
                    : { rotate: 0 }
                }
                transition={
                  loading && !reducedMotion
                    ? { duration: 1, ease: "linear", repeat: Infinity }
                    : { duration: 0.2 }
                }
              >
                {loading ? (
                  <LoaderCircle
                    size={16}
                    strokeWidth={2.2}
                  />
                ) : (
                    <ArrowRight size={16} strokeWidth={2.4} />
                )}
              </motion.span>
            </motion.button>
          </motion.form>
        </motion.div>
        <motion.div
          className="auth-panel auth-panel--visual"
          variants={rightPanelVariants}
        >
          <motion.div className="auth-visual" variants={slideVariants}>
            <motion.div
              className="auth-visual__flare"
              animate={
                reducedMotion
                  ? undefined
                  : {
                      scale: [1, 1.06, 0.98, 1],
                      x: [0, 18, -8, 0],
                      y: [0, -10, 6, 0],
                      opacity: [0.88, 1, 0.9, 0.88],
                    }
              }
              transition={
                reducedMotion
                  ? undefined
                  : {
                      duration: 11,
                      ease: "easeInOut",
                      repeat: Infinity,
                    }
              }
            />
            <motion.p className="auth-visual__eyebrow" variants={slideVariants}>
              ProfDNK platform
            </motion.p>
            <motion.h2 className="auth-visual__title" variants={slideVariants}>
              Платформа
              <br />
              <span className="auth-visual__title-line">профориентации и</span>
              <br />
              психодиагностики
            </motion.h2>
            <motion.p className="auth-visual__text" variants={slideVariants}>
              Создание методик, клиентские сессии, результаты и отчёты в одном
              месте.
            </motion.p>
          </motion.div>
        </motion.div>
      </motion.div>
    </motion.section>
  );
}
