import { Ionicons } from '@expo/vector-icons';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { theme } from '../constants/theme';

type Props = {
  title: string;
  subtitle?: string;
  onActionPress?: () => void;
  actionIcon?: keyof typeof Ionicons.glyphMap;
};

export const AppHeader = ({ title, subtitle, onActionPress, actionIcon = 'menu' }: Props) => {
  return (
    <View style={styles.header}>
      <View>
        <Text style={styles.logo}>Eventor</Text>
        <Text style={styles.title}>{title}</Text>
        {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
      </View>
      {onActionPress ? (
        <Pressable onPress={onActionPress} style={styles.iconButton}>
          <Ionicons name={actionIcon} size={24} color={theme.colors.text} />
        </Pressable>
      ) : null}
    </View>
  );
};

const styles = StyleSheet.create({
  header: {
    marginTop: theme.spacing.sm,
    marginBottom: theme.spacing.sm,
    paddingHorizontal: theme.spacing.md,
    paddingVertical: theme.spacing.sm,
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  logo: {
    fontSize: 14,
    color: theme.colors.primary,
    fontWeight: '700',
    letterSpacing: 0.8,
    textTransform: 'uppercase',
  },
  title: {
    fontSize: 20,
    color: theme.colors.text,
    fontWeight: '700',
  },
  subtitle: {
    marginTop: 2,
    fontSize: 13,
    color: theme.colors.muted,
  },
  iconButton: {
    width: 44,
    height: 44,
    borderRadius: 12,
    borderColor: theme.colors.border,
    borderWidth: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: theme.colors.card,
  },
});
