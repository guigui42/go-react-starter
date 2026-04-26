import { useUserPreferencesContext } from '@/contexts/UserPreferencesContext';
import { ActionIcon, useMantineColorScheme } from '@mantine/core';
import { IconMoon, IconSun } from '@tabler/icons-react';

export function ColorSchemeToggle() {
  const { colorScheme: mantineColorScheme } = useMantineColorScheme();
  const { setColorScheme } = useUserPreferencesContext();

  // Always toggle based on Mantine's resolved scheme (what's actually displayed)
  // This handles 'auto' mode correctly - if system is dark, clicking switches to light
  const handleToggle = () => {
    const newScheme = mantineColorScheme === 'dark' ? 'light' : 'dark';
    setColorScheme(newScheme);
  };

  return (
    <ActionIcon
      onClick={handleToggle}
      variant="default"
      size="lg"
      aria-label="Toggle color scheme"
    >
      {mantineColorScheme === 'dark' ? <IconSun size={20} /> : <IconMoon size={20} />}
    </ActionIcon>
  );
}
