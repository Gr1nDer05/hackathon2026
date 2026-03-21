function firstNonEmpty(...values) {
  for (const value of values) {
    if (String(value || "").trim()) {
      return String(value).trim();
    }
  }

  return "";
}

function firstBoolean(...values) {
  for (const value of values) {
    if (typeof value === "boolean") {
      return value;
    }
  }

  return undefined;
}

function pickObject(...values) {
  return (
    values.find(
      (value) => value && typeof value === "object" && !Array.isArray(value),
    ) || null
  );
}

export function normalizePublicTestAuthor(test) {
  const source = pickObject(
    test?.psychologist,
    test?.author,
    test?.owner,
    test?.created_by_user,
    test?.created_by,
    test?.specialist,
  );
  const account = pickObject(source?.user, source?.account, source);
  const profile = pickObject(
    source?.profile,
    test?.psychologist_profile,
    test?.author_profile,
    account?.profile,
  );
  const card = pickObject(
    source?.card,
    test?.psychologist_card,
    test?.author_card,
    account?.card,
  );

  const normalized = {
    ...account,
    id:
      account?.id ||
      source?.id ||
      test?.psychologist_id ||
      test?.author_id ||
      test?.created_by_user_id ||
      null,
    full_name: firstNonEmpty(
      account?.full_name,
      account?.name,
      source?.full_name,
      source?.name,
      test?.psychologist_full_name,
      test?.psychologist_name,
      test?.author_full_name,
      test?.author_name,
    ),
    email: firstNonEmpty(
      account?.email,
      source?.email,
      card?.contact_email,
      test?.psychologist_email,
      test?.author_email,
    ),
    profile: {
      ...profile,
      specialization: firstNonEmpty(
        profile?.specialization,
        source?.specialization,
        card?.specialization,
        test?.psychologist_specialization,
        test?.author_specialization,
      ),
      city: firstNonEmpty(
        profile?.city,
        source?.city,
        card?.city,
        test?.psychologist_city,
        test?.author_city,
      ),
      about: firstNonEmpty(
        profile?.about,
        source?.about,
        card?.about,
        test?.psychologist_about,
        test?.author_about,
      ),
      education: firstNonEmpty(
        profile?.education,
        source?.education,
        test?.psychologist_education,
        test?.author_education,
      ),
      methods: firstNonEmpty(
        profile?.methods,
        source?.methods,
        test?.psychologist_methods,
        test?.author_methods,
      ),
      timezone: firstNonEmpty(
        profile?.timezone,
        source?.timezone,
        test?.psychologist_timezone,
        test?.author_timezone,
      ),
      experience_years:
        profile?.experience_years ??
        source?.experience_years ??
        test?.psychologist_experience_years ??
        test?.author_experience_years ??
        null,
      is_public:
        firstBoolean(profile?.is_public, source?.is_public, card?.is_public) ??
        true,
    },
    card: {
      ...card,
      headline: firstNonEmpty(
        card?.headline,
        source?.headline,
        test?.psychologist_headline,
        test?.author_headline,
      ),
      short_bio: firstNonEmpty(
        card?.short_bio,
        source?.short_bio,
        test?.psychologist_short_bio,
        test?.author_short_bio,
      ),
      contact_email: firstNonEmpty(
        card?.contact_email,
        source?.contact_email,
        account?.email,
        source?.email,
        test?.psychologist_email,
        test?.author_email,
      ),
      contact_phone: firstNonEmpty(
        card?.contact_phone,
        source?.contact_phone,
        source?.phone,
        account?.phone,
        test?.psychologist_phone,
        test?.author_phone,
      ),
      telegram: firstNonEmpty(
        card?.telegram,
        source?.telegram,
        test?.psychologist_telegram,
        test?.author_telegram,
      ),
      website: firstNonEmpty(
        card?.website,
        source?.website,
        test?.psychologist_website,
        test?.author_website,
      ),
      online_available:
        firstBoolean(
          card?.online_available,
          source?.online_available,
          test?.psychologist_online_available,
          test?.author_online_available,
        ) ?? false,
      offline_available:
        firstBoolean(
          card?.offline_available,
          source?.offline_available,
          test?.psychologist_offline_available,
          test?.author_offline_available,
        ) ?? false,
    },
  };

  const hasContent = [
    normalized.full_name,
    normalized.email,
    normalized.profile.specialization,
    normalized.profile.city,
    normalized.profile.about,
    normalized.card.headline,
    normalized.card.short_bio,
    normalized.card.contact_phone,
    normalized.card.telegram,
    normalized.card.website,
  ].some((value) => String(value || "").trim());

  return hasContent ? normalized : null;
}

export function hasPublicTestAuthorAccess(test) {
  return Boolean(
    normalizePublicTestAuthor(test) ||
      test?.created_by_user_id ||
      test?.psychologist_id ||
      test?.author_id ||
      test?.psychologist ||
      test?.author,
  );
}
