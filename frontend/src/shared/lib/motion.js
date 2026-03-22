export const MOTION_EASE = [0.22, 1, 0.36, 1];

export function createRevealContainer(
  reducedMotion,
  {
    staggerChildren = 0.08,
    delayChildren = 0,
    duration = 0.3,
  } = {},
) {
  return {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        duration: reducedMotion ? 0.01 : duration,
        ease: MOTION_EASE,
        staggerChildren: reducedMotion ? 0 : staggerChildren,
        delayChildren: reducedMotion ? 0 : delayChildren,
      },
    },
  };
}

export function createFadeMove(
  reducedMotion,
  { axis = "y", distance = 18, scale = 1 } = {},
) {
  const hidden = {
    opacity: 0,
    [axis]: reducedMotion ? 0 : distance,
  };

  if (scale !== 1) {
    hidden.scale = reducedMotion ? 1 : scale;
  }

  return {
    hidden,
    visible: {
      opacity: 1,
      [axis]: 0,
      ...(scale !== 1 ? { scale: 1 } : {}),
      transition: {
        duration: reducedMotion ? 0.01 : 0.55,
        ease: MOTION_EASE,
      },
    },
  };
}
