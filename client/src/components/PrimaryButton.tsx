import { ActivityIndicator, Pressable, StyleSheet, Text } from 'react-native';
import { theme } from '../constants/theme';

type Props = {
  title: string;
  onPress: () => void;
  disabled?: boolean;
  loading?: boolean;
  type?: 'primary' | 'outline';
};

export const PrimaryButton = ({ title, onPress, disabled, loading, type = 'primary' }: Props) => {
  const isOutline = type === 'outline';
  return (
    <Pressable
      disabled={disabled || loading}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        isOutline ? styles.outlineButton : styles.primaryButton,
        (disabled || loading) && styles.disabled,
        pressed && styles.pressed,
      ]}
    >
      {loading ? (
        <ActivityIndicator color={isOutline ? theme.colors.primary : '#fff'} />
      ) : (
        <Text style={[styles.title, isOutline ? styles.outlineTitle : styles.primaryTitle]}>{title}</Text>
      )}
    </Pressable>
  );
};

const styles = StyleSheet.create({
  button: {
    height: 46,
    borderRadius: theme.radius.sm,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
  },
  primaryButton: {
    borderColor: theme.colors.primary,
    backgroundColor: theme.colors.primary,
  },
  outlineButton: {
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
  },
  title: {
    fontSize: 14,
    fontWeight: '700',
  },
  primaryTitle: {
    color: '#fff',
  },
  outlineTitle: {
    color: theme.colors.primary,
  },
  disabled: {
    opacity: 0.6,
  },
  pressed: {
    transform: [{ scale: 0.98 }],
  },
});
