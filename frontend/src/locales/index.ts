import zhCN from './zh-CN';
import enUS from './en-US';

export const locales = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

export type LocaleCode = keyof typeof locales;

export const localeNames: Record<LocaleCode, string> = {
  'zh-CN': '中文',
  'en-US': 'EN',
};

export { zhCN, enUS };
