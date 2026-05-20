import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { ScreenContainer } from '../components/ScreenContainer';
import { AppHeader } from '../components/AppHeader';
import { AnimatedEntry } from '../components/AnimatedEntry';
import { FormInput } from '../components/FormInput';
import { PrimaryButton } from '../components/PrimaryButton';
import { theme } from '../constants/theme';
import { RootStackParamList } from '../navigation/types';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { loginRequest, registerRequest } from '../redux/slices/authSlice';
import { UserRole } from '../types/models';

type Props = NativeStackScreenProps<RootStackParamList, 'Auth'>;

export const AuthPage = ({ navigation }: Props) => {
  const dispatch = useAppDispatch();
  const auth = useAppSelector((state) => state.auth);
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [name, setName] = useState('');
  const [city, setCity] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  

  const onSubmit = () => {
    if (mode === 'login') {
      dispatch(loginRequest({ email, password }));
      return;
    }
    dispatch(registerRequest({ name, city, email, password }));
  };

  return (
    <ScreenContainer>
      <AppHeader title="Добро пожаловать" subtitle="Регистрация и авторизация" />

      <AnimatedEntry delay={50}>
        <View style={styles.panel}>
          <View style={styles.modeRow}>
            <Pressable style={[styles.modeButton, mode === 'login' && styles.modeButtonActive]} onPress={() => setMode('login')}>
              <Text style={[styles.modeText, mode === 'login' && styles.modeTextActive]}>Вход</Text>
            </Pressable>
            <Pressable
              style={[styles.modeButton, mode === 'register' && styles.modeButtonActive]}
              onPress={() => setMode('register')}
            >
              <Text style={[styles.modeText, mode === 'register' && styles.modeTextActive]}>Регистрация</Text>
            </Pressable>
          </View>

          {mode === 'register' ? <FormInput label="Имя" value={name} onChangeText={setName} placeholder="Ваше имя" /> : null}
          {mode === 'register' ? <FormInput label="Город" value={city} onChangeText={setCity} placeholder="Ваш город" /> : null}
          <FormInput label="Почта" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" placeholder="email@example.com" />
          <FormInput label="Пароль" value={password} onChangeText={setPassword} secureTextEntry />

          {/* role selection removed: registrations always create regular users */}

          {auth.error ? <Text style={styles.error}>{auth.error}</Text> : null}

          <PrimaryButton
            title={mode === 'login' ? 'Войти' : 'Создать аккаунт'}
            onPress={onSubmit}
            loading={auth.loading}
            disabled={!email || !password || (mode === 'register' && (!name || !city))}
          />

          <PrimaryButton title="Смотреть мероприятия" type="outline" onPress={() => navigation.navigate('Events')} />
        </View>
      </AnimatedEntry>
    </ScreenContainer>
  );
};

const styles = StyleSheet.create({
  panel: {
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surface,
    padding: theme.spacing.md,
    gap: theme.spacing.sm,
  },
  modeRow: {
    flexDirection: 'row',
    backgroundColor: theme.colors.card,
    borderRadius: theme.radius.md,
    padding: 4,
    gap: 6,
  },
  modeButton: {
    flex: 1,
    height: 38,
    borderRadius: 10,
    alignItems: 'center',
    justifyContent: 'center',
  },
  modeButtonActive: {
    backgroundColor: theme.colors.surface,
    borderWidth: 1,
    borderColor: theme.colors.border,
  },
  modeText: {
    fontSize: 13,
    color: theme.colors.muted,
    fontWeight: '600',
  },
  modeTextActive: {
    color: theme.colors.text,
  },
  roleRow: {
    flexDirection: 'row',
    gap: theme.spacing.sm,
  },
  roleButton: {
    flex: 1,
    height: 40,
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.radius.sm,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: theme.colors.card,
  },
  roleButtonActive: {
    borderColor: theme.colors.primary,
    backgroundColor: theme.colors.primarySoft,
  },
  roleText: {
    fontSize: 13,
    color: theme.colors.muted,
    fontWeight: '600',
  },
  roleTextActive: {
    color: theme.colors.primary,
  },
  error: {
    color: theme.colors.danger,
    fontSize: 13,
  },
});
