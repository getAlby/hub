/** @type {import('tailwindcss').Config} */
import defaultTheme from "tailwindcss/defaultTheme";
import forms from "@tailwindcss/forms";
import aspectRatio from "@tailwindcss/aspect-ratio";
import typography from "@tailwindcss/typography";

// NOTE: copied from:
//   https://github.com/getAlby/lightning-browser-extension/blob/1bab35feb58859d0786a9c2807a038395dd87fca/tailwind.config.js#L4-L21
function lighten(color, percent) {
  var num = parseInt(color.replace("#", ""), 16),
    amt = Math.round(2.55 * percent),
    R = (num >> 16) + amt,
    B = ((num >> 8) & 0x00ff) + amt,
    G = (num & 0x0000ff) + amt;
  return (
    "#" +
    (
      0x1000000 +
      (R < 255 ? (R < 1 ? 0 : R) : 255) * 0x10000 +
      (B < 255 ? (B < 1 ? 0 : B) : 255) * 0x100 +
      (G < 255 ? (G < 1 ? 0 : G) : 255)
    )
      .toString(16)
      .slice(1)
  );
}

const surfaceColor = "#121212";

export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        mono: defaultTheme.fontFamily.mono,
        sans: ["Inter var", ...defaultTheme.fontFamily.sans],
      },
      colors: {
        "alby-grey": "#272828",
        "alby-yellow": "#FFDE6E",
        "alby-yellow-light": "#FDF0D5",
        bitcoin: "#F7931A",
        "cold-grey": "#C5C7C8",
        light: "#F4F4F4",
        primary: "#9333ea",
        "warm-grey": "#D2D2D1",

        // Material Design Surface Colors
        "surface-00dp": surfaceColor,
        "surface-01dp": lighten(surfaceColor, 5),
        "surface-02dp": lighten(surfaceColor, 7),
        "surface-03dp": lighten(surfaceColor, 8),
        "surface-04dp": lighten(surfaceColor, 9),
        "surface-06dp": lighten(surfaceColor, 11),
        "surface-08dp": lighten(surfaceColor, 12),
        "surface-12dp": lighten(surfaceColor, 14),
        "surface-16dp": lighten(surfaceColor, 15),
        "surface-24dp": lighten(surfaceColor, 16),
      },
      borderRadius: {
        "4xl": "1.25rem",
        30: "30px",
      },
      animation: {
        wiggle: "wiggle 0.3s ease-in-out",
      },
      keyframes: {
        wiggle: {
          "0%, 100%": { transform: "scale(1.0)" },
          "50%": { transform: "scale(1.03)" },
        },
      },
      blur: {
        xs: "2px",
      },
    },
    backgroundImage: () => ({
      "primary-gradient": `linear-gradient(180deg, #FFDE6E 63%, #F8C455 95%)`,
      "primary-gradient-hover": `linear-gradient(180deg, #F2D369 63%, #ECBA51 95%)`,
    }),
    fontFamily: {
      headline: [
        '"Work Sans"',
        '"Inter var"',
        "Helvetica",
        "Arial",
        "sans-serif",
      ],
      // required for Chrome, where e.g. ❤️ in Inter is displayed as a black heart :(
      emoji: ["sans-serif"],
    },
  },
  plugins: [forms, aspectRatio, typography],
};
