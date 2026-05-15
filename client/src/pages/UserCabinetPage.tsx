import * as ImagePicker from 'expo-image-picker';
import { useEffect, useState } from 'react';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { Alert, StyleSheet, Text, View } from 'react-native';
import { AnimatedEntry } from '../components/AnimatedEntry';
import { AppHeader } from '../components/AppHeader';
import { EventCard } from '../components/EventCard';
import { FormInput } from '../components/FormInput';
import { PrimaryButton } from '../components/PrimaryButton';
import { ProfileAvatar } from '../components/ProfileAvatar';
import { ScreenContainer } from '../components/ScreenContainer';
import { RootStackParamList } from '../navigation/types';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { logout } from '../redux/slices/authSlice';
import { registerForEventRequest } from '../redux/slices/eventsSlice';
import { updateAvatarRequest, updateProfileRequest } from '../redux/slices/profileSlice';
import { theme } from '../constants/theme';

type Props = NativeStackScreenProps<RootStackParamList, 'UserCabinet'>;

export const UserCabinetPage = ({ navigation }: Props) => {
  const dispatch = useAppDispatch();
  const profile = useAppSelector((state) => state.profile.profile);
  const events = useAppSelector((state) => state.events.list);

  const [name, setName] = useState('');
  const [city, setCity] = useState('');
  const [about, setAbout] = useState('');
  const [faculty, setFaculty] = useState('');
  const [course, setCourse] = useState('');

  useEffect(() => {
    if (!profile) {
      return;
    }
    setName(profile.name);
    setCity(profile.city);
    setAbout(profile.about);
    setFaculty(profile.faculty);
    setCourse(profile.course);
  }, [profile]);

  if (!profile) {
    return null;
  }

  const myEvents = events.filter((eventItem) => eventItem.attendees.includes(profile.id));

  const pickAvatar = async () => {
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      Alert.alert('Нет доступа', 'Разрешите доступ к фото для выбора аватара');
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      quality: 0.8,
    });

    if (result.canceled || !result.assets[0]?.uri) {
      return;
    }

    dispatch(updateAvatarRequest(result.assets[0].uri));
  };

  return (
    <ScreenContainer>
      <AppHeader title="Личный кабинет" subtitle="Пользователь" onActionPress={() => dispatch(logout())} actionIcon="log-out-outline" />

      <AnimatedEntry>
        <View style={styles.card}>
          <View style={styles.centered}>
            <ProfileAvatar uri={profile.avatarUri} />
          </View>

          <PrimaryButton title="Обновить аватар" type="outline" onPress={pickAvatar} />

          <FormInput label="Имя" value={name} onChangeText={setName} />
          <FormInput label="Город" value={city} onChangeText={setCity} />
          <FormInput label="О себе" value={about} onChangeText={setAbout} />
          <FormInput label="Факультет" value={faculty} onChangeText={setFaculty} />
          <FormInput label="Курс" value={course} onChangeText={setCourse} keyboardType="number-pad" />

          <PrimaryButton
            title="Обновить профиль"
            onPress={() => dispatch(updateProfileRequest({ name, city, about, faculty, course }))}
          />
        </View>
      </AnimatedEntry>

      <PrimaryButton title="Открыть страницу мероприятий" type="outline" onPress={() => navigation.navigate('Events')} />

      <Text style={styles.sectionTitle}>Мои мероприятия</Text>
      {myEvents.length === 0 ? <Text style={styles.empty}>Вы пока не записаны на мероприятия</Text> : null}
      {myEvents.map((eventItem) => (
        <EventCard
          key={eventItem.id}
          eventItem={eventItem}
          isRegistered
          onPress={() => navigation.navigate('EventDetails', { eventId: eventItem.id })}
          onToggleRegistration={() => dispatch(registerForEventRequest(eventItem.id))}
        />
      ))}
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  card: {
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    padding: theme.spacing.md,
    gap: theme.spacing.sm,
  },
  centered: {
    alignItems: 'center',
    marginBottom: theme.spacing.xs,
  },
  sectionTitle: {
    color: theme.colors.text,
    fontSize: 18,
    fontWeight: '700',
    marginTop: 2,
  },
  empty: {
    color: theme.colors.muted,
    fontSize: 13,
  },
});
