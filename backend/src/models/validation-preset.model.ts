import { PrismaClient } from '@prisma/client';
import { CreatePresetInput, UpdatePresetInput, ValidationPresetRecord } from '../types/preset.types';

const prisma = new PrismaClient();

export const ValidationPresetModel = {
  async countByUser(userId: string): Promise<number> {
    return prisma.validationPreset.count({ where: { userId } });
  },

  async create(input: CreatePresetInput): Promise<ValidationPresetRecord> {
    return prisma.validationPreset.create({ data: input });
  },

  async findAllByUser(userId: string): Promise<ValidationPresetRecord[]> {
    return prisma.validationPreset.findMany({
      where: { userId },
      orderBy: { createdAt: 'asc' },
    });
  },

  async findByIdAndUser(id: string, userId: string): Promise<ValidationPresetRecord | null> {
    return prisma.validationPreset.findFirst({ where: { id, userId } });
  },

  async update(id: string, userId: string, input: UpdatePresetInput): Promise<ValidationPresetRecord> {
    return prisma.validationPreset.update({ where: { id }, data: input });
  },

  async delete(id: string, userId: string): Promise<void> {
    await prisma.validationPreset.delete({ where: { id } });
  },
};
