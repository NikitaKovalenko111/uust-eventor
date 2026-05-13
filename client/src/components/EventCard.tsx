import { Pressable, StyleSheet, Text, View } from 'react-native';
import { EventItem } from '../types/models';
import { theme } from '../constants/theme';
import { PrimaryButton } from './PrimaryButton';

type Props = {
  eventItem: EventItem;
  isRegistered: boolean;
  onPress: () => void;
  onToggleRegistration?: () => void;
};

export const EventCard = ({ eventItem, isRegistered, onPress, onToggleRegistration }: Props) => {
  return (
    <Pressable onPress={onPress} style={styles.card}>
      <View style={styles.row}>
        <Text style={styles.title}>{eventItem.title}</Text>
        <Text style={styles.date}>{eventItem.date}</Text>
      </View>
      <Text style={styles.description}>{eventItem.shortDescription}</Text>
      <Text style={styles.meta}>{eventItem.location}</Text>
      <Text style={styles.tags}>{eventItem.tags.join(' • ')}</Text>
      {onToggleRegistration ? (
        <View style={styles.buttonWrap}>
          <PrimaryButton
            title={isRegistered ? 'Отменить регистрацию' : 'Записаться'}
            onPress={onToggleRegistration}
            type={isRegistered ? 'outline' : 'primary'}
          />
        </View>
      ) : null}
    </Pressable>
  );
};

const styles = StyleSheet.create({
  card: {
    borderRadius: theme.radius.md,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    padding: theme.spacing.md,
    gap: 8,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  title: {
    flex: 1,
    color: theme.colors.text,
    fontSize: 16,
    fontWeight: '700',
  },
  date: {
    color: theme.colors.primary,
    fontSize: 13,
    fontWeight: '600',
  },
  description: {
    color: theme.colors.text,
    fontSize: 14,
  },
  meta: {
    color: theme.colors.muted,
    fontSize: 12,
  },
  tags: {
    color: theme.colors.accent,
    fontSize: 12,
    fontWeight: '600',
  },
  buttonWrap: {
    marginTop: 6,
  },
});
