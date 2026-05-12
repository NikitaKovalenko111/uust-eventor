import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { Image, StyleSheet, Text, View } from 'react-native';
import { AppHeader } from '../components/AppHeader';
import { ScreenContainer } from '../components/ScreenContainer';
import { theme } from '../constants/theme';
import { RootStackParamList } from '../navigation/types';
import { useAppSelector } from '../redux/hooks';

type Props = NativeStackScreenProps<RootStackParamList, 'EventDetails'>;

export const EventDetailsPage = ({ route, navigation }: Props) => {
  const eventItem = useAppSelector((state) => state.events.list.find((event) => event.id === route.params.eventId));

  if (!eventItem) {
    return (
      <ScreenContainer>
        <AppHeader title="Мероприятие" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />
        <Text style={styles.empty}>Мероприятие не найдено</Text>
      </ScreenContainer>
    );
  }

  return (
    <ScreenContainer>
      <AppHeader title="Страница мероприятия" onActionPress={() => navigation.goBack()} actionIcon="arrow-back" />

      <View style={styles.photoBlock}>
        <Image source={{ uri: eventItem.imageUri }} style={styles.image} />
      </View>

      <Text style={styles.title}>{eventItem.title}</Text>
      <Text style={styles.description}>{eventItem.description}</Text>
      <Text style={styles.meta}>Дата: {eventItem.date}</Text>
      <Text style={styles.meta}>Место: {eventItem.location}</Text>

      <View style={styles.discussionBlock}>
        <Text style={styles.discussionTitle}>Блок обсуждения мероприятия</Text>
        <Text style={styles.discussionText}>Здесь можно добавить чат, комментарии и вопросы к организаторам.</Text>
      </View>
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  photoBlock: {
    borderRadius: theme.radius.md,
    borderWidth: 1,
    borderColor: theme.colors.border,
    overflow: 'hidden',
    backgroundColor: theme.colors.surface,
  },
  image: {
    width: '100%',
    height: 190,
  },
  title: {
    marginTop: 4,
    color: theme.colors.text,
    fontSize: 20,
    fontWeight: '700',
  },
  description: {
    color: theme.colors.text,
    fontSize: 14,
    lineHeight: 22,
  },
  meta: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  discussionBlock: {
    marginTop: theme.spacing.sm,
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.radius.md,
    padding: theme.spacing.md,
    backgroundColor: theme.colors.surface,
    gap: 8,
  },
  discussionTitle: {
    color: theme.colors.text,
    fontWeight: '700',
    fontSize: 15,
  },
  discussionText: {
    color: theme.colors.muted,
    fontSize: 13,
  },
  empty: {
    marginTop: 24,
    textAlign: 'center',
    color: theme.colors.muted,
  },
});
